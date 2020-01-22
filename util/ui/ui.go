// Package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
//
// This package is explicitly designed for the CF CLI and is *not* to be used
// by any package outside of the commands package.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/fatih/color"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/vito/go-interact/interact"
)

var realExiter exiterFunc = os.Exit

var realInteract interactorFunc = func(prompt string, choices ...interact.Choice) Resolver {
	return &interactionWrapper{
		interact.NewInteraction(prompt, choices...),
	}
}

//go:generate counterfeiter . Interactor

// Interactor hides interact.NewInteraction for testing purposes
type Interactor interface {
	NewInteraction(prompt string, choices ...interact.Choice) Resolver
}

type interactorFunc func(prompt string, choices ...interact.Choice) Resolver

func (f interactorFunc) NewInteraction(prompt string, choices ...interact.Choice) Resolver {
	return f(prompt, choices...)
}

type interactionWrapper struct {
	interact.Interaction
}

func (w *interactionWrapper) SetIn(in io.Reader) {
	w.Input = in
}

func (w *interactionWrapper) SetOut(o io.Writer) {
	w.Output = o
}

//go:generate counterfeiter . Exiter

// Exiter hides os.Exit for testing purposes
type Exiter interface {
	Exit(code int)
}

type exiterFunc func(int)

func (f exiterFunc) Exit(code int) {
	f(code)
}

// UI is interface to interact with the user
type UI struct {
	// In is the input buffer
	In io.Reader
	// Out is the output buffer
	Out io.Writer
	// OutForInteration is the output buffer when working with go-interact. When
	// working with Windows, color.Output does not work with TTY detection. So
	// real STDOUT is required or go-interact will not properly work.
	OutForInteration io.Writer
	// Err is the error buffer
	Err io.Writer

	colorEnabled configv3.ColorSetting
	translate    TranslateFunc
	Exiter       Exiter

	terminalLock *sync.Mutex
	fileLock     *sync.Mutex

	Interactor Interactor

	IsTTY         bool
	TerminalWidth int

	TimezoneLocation *time.Location

	deferred []string
}

// NewUI will return a UI object where Out is set to STDOUT, In is set to
// STDIN, and Err is set to STDERR
func NewUI(config Config) (*UI, error) {
	translateFunc, err := GetTranslationFunc(config)
	if err != nil {
		return nil, err
	}

	location := time.Now().Location()

	return &UI{
		In:               os.Stdin,
		Out:              color.Output,
		OutForInteration: os.Stdout,
		Err:              os.Stderr,
		colorEnabled:     config.ColorEnabled(),
		translate:        translateFunc,
		terminalLock:     &sync.Mutex{},
		Exiter:           realExiter,
		fileLock:         &sync.Mutex{},
		Interactor:       realInteract,
		IsTTY:            config.IsTTY(),
		TerminalWidth:    config.TerminalWidth(),
		TimezoneLocation: location,
	}, nil
}

// NewPluginUI will return a UI object where OUT and ERR are customizable.
func NewPluginUI(config Config, outBuffer io.Writer, errBuffer io.Writer) (*UI, error) {
	translateFunc, translationError := GetTranslationFunc(config)
	if translationError != nil {
		return nil, translationError
	}

	location := time.Now().Location()

	return &UI{
		In:               nil,
		Out:              outBuffer,
		OutForInteration: outBuffer,
		Err:              errBuffer,
		colorEnabled:     configv3.ColorDisabled,
		translate:        translateFunc,
		terminalLock:     &sync.Mutex{},
		Exiter:           realExiter,
		fileLock:         &sync.Mutex{},
		Interactor:       realInteract,
		IsTTY:            config.IsTTY(),
		TerminalWidth:    config.TerminalWidth(),
		TimezoneLocation: location,
	}, nil
}

// NewTestUI will return a UI object where Out, In, and Err are customizable,
// and colors are disabled
func NewTestUI(in io.Reader, out io.Writer, err io.Writer) *UI {
	translationFunc, translateErr := generateTranslationFunc([]byte("[]"))
	if translateErr != nil {
		panic(translateErr)
	}

	return &UI{
		In:               in,
		Out:              out,
		OutForInteration: out,
		Err:              err,
		Exiter:           realExiter,
		colorEnabled:     configv3.ColorDisabled,
		translate:        translationFunc,
		Interactor:       realInteract,
		terminalLock:     &sync.Mutex{},
		fileLock:         &sync.Mutex{},
		TimezoneLocation: time.UTC,
	}
}

// DeferText translates the template, substitutes in templateValues, and
// Enqueues the output to be presented later via FlushDeferred. Only the first
// map in templateValues is used.
func (ui *UI) DeferText(template string, templateValues ...map[string]interface{}) {
	s := fmt.Sprintf("%s\n", ui.TranslateText(template, templateValues...))
	ui.deferred = append(ui.deferred, s)
}

func (ui *UI) DisplayDeprecationWarning() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Err, "Deprecation warning: This command has been deprecated. This feature will be removed in the future.\n")
}

// DisplayError outputs the translated error message to ui.Err if the error
// satisfies TranslatableError, otherwise it outputs the original error message
// to ui.Err. It also outputs "FAILED" in bold red to ui.Out.
func (ui *UI) DisplayError(err error) {
	var errMsg string
	if translatableError, ok := err.(translatableerror.TranslatableError); ok {
		errMsg = translatableError.Translate(ui.translate)
	} else {
		errMsg = err.Error()
	}
	fmt.Fprintf(ui.Err, "%s\n", errMsg)

	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.modifyColor(ui.TranslateText("FAILED"), color.New(color.FgRed, color.Bold)))
}

func (ui *UI) DisplayFileDeprecationWarning() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Err, "Deprecation warning: This command has been deprecated and will be removed in the future. For similar functionality, please use the `cf ssh` command instead.\n")
}

// DisplayHeader translates the header, bolds and adds the default color to the
// header, and outputs the result to ui.Out.
func (ui *UI) DisplayHeader(text string) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.modifyColor(ui.TranslateText(text), color.New(color.Bold)))
}

// DisplayNewline outputs a newline to UI.Out.
func (ui *UI) DisplayNewline() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "\n")
}

// DisplayOK outputs a bold green translated "OK" to UI.Out.
func (ui *UI) DisplayOK() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n\n", ui.modifyColor(ui.TranslateText("OK"), color.New(color.FgGreen, color.Bold)))
}

// DisplayText translates the template, substitutes in templateValues, and
// outputs the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayText(template string, templateValues ...map[string]interface{}) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, templateValues...))
}

// DisplayTextWithBold translates the template, bolds the templateValues,
// substitutes templateValues into the template, and outputs
// the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayTextWithBold(template string, templateValues ...map[string]interface{}) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	firstTemplateValues := getFirstSet(templateValues)
	for key, value := range firstTemplateValues {
		firstTemplateValues[key] = ui.modifyColor(fmt.Sprint(value), color.New(color.Bold))
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, firstTemplateValues))
}

// DisplayTextWithFlavor translates the template, bolds and adds cyan color to
// templateValues, substitutes templateValues into the template, and outputs
// the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayTextWithFlavor(template string, templateValues ...map[string]interface{}) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	firstTemplateValues := getFirstSet(templateValues)
	for key, value := range firstTemplateValues {
		firstTemplateValues[key] = ui.modifyColor(fmt.Sprint(value), color.New(color.FgCyan, color.Bold))
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, firstTemplateValues))
}

// FlushDeferred displays text previously deferred (using DeferText) to the UI's
// `Out`.
func (ui *UI) FlushDeferred() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	for _, s := range ui.deferred {
		fmt.Fprint(ui.Out, s)
	}
	ui.deferred = []string{}
}

// GetErr returns the error writer.
func (ui *UI) GetErr() io.Writer {
	return ui.Err
}

// GetIn returns the input reader.
func (ui *UI) GetIn() io.Reader {
	return ui.In
}

// GetOut returns the output writer. Same as `Writer`.
func (ui *UI) GetOut() io.Writer {
	return ui.Out
}

// TranslateText passes the template through an internationalization function
// to translate it to a pre-configured language, and returns the template with
// templateValues substituted in. Only the first map in templateValues is used.
func (ui *UI) TranslateText(template string, templateValues ...map[string]interface{}) string {
	return ui.translate(template, getFirstSet(templateValues))
}

// UserFriendlyDate converts the time to UTC and then formats it to ISO8601.
func (ui *UI) UserFriendlyDate(input time.Time) string {
	return input.Local().Format("Mon 02 Jan 15:04:05 MST 2006")
}

// Writer returns the output writer. Same as `GetOut`.
func (ui *UI) Writer() io.Writer {
	return ui.Out
}

func (ui *UI) displayWrappingTableWithWidth(prefix string, table [][]string, padding int) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var columnPadding []int

	rows := len(table)
	columns := len(table[0])

	for col := 0; col < columns-1; col++ {
		var max int
		for row := 0; row < rows; row++ {
			if strLen := runewidth.StringWidth(table[row][col]); max < strLen {
				max = strLen
			}
		}
		columnPadding = append(columnPadding, max+padding)
	}

	spilloverPadding := len(prefix) + sum(columnPadding)
	lastColumnWidth := ui.TerminalWidth - spilloverPadding

	for row := 0; row < rows; row++ {
		fmt.Fprint(ui.Out, prefix)

		// for all columns except last, add cell value and padding
		for col := 0; col < columns-1; col++ {
			var addedPadding int
			if col+1 != columns {
				addedPadding = columnPadding[col] - runewidth.StringWidth(table[row][col])
			}
			fmt.Fprintf(ui.Out, "%s%s", table[row][col], strings.Repeat(" ", addedPadding))
		}

		// for last column, add each word individually. If the added word would make the column exceed terminal width, create a new line and add padding
		words := strings.Split(table[row][columns-1], " ")
		currentWidth := 0

		for _, word := range words {
			wordWidth := runewidth.StringWidth(word)
			switch {
			case currentWidth == 0:
				currentWidth = wordWidth
				fmt.Fprintf(ui.Out, "%s", word)
			case (wordWidth + 1 + currentWidth) > lastColumnWidth:
				fmt.Fprintf(ui.Out, "\n%s%s", strings.Repeat(" ", spilloverPadding), word)
				currentWidth = wordWidth
			default:
				fmt.Fprintf(ui.Out, " %s", word)
				currentWidth += wordWidth + 1
			}
		}

		fmt.Fprintf(ui.Out, "\n")
	}
}

func (ui *UI) modifyColor(text string, colorPrinter *color.Color) string {
	if len(text) == 0 {
		return text
	}

	switch ui.colorEnabled {
	case configv3.ColorEnabled:
		colorPrinter.EnableColor()
	case configv3.ColorDisabled:
		colorPrinter.DisableColor()
	}

	return colorPrinter.SprintFunc()(text)
}

// getFirstSet returns the first map if 1 or more maps are provided. Otherwise
// it returns the empty map.
func getFirstSet(list []map[string]interface{}) map[string]interface{} {
	if len(list) == 0 {
		return map[string]interface{}{}
	}
	return list[0]
}

func sum(intSlice []int) int {
	sum := 0

	for _, i := range intSlice {
		sum += i
	}

	return sum
}
