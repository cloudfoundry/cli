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
	"text/tabwriter"

	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/fatih/color"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/vito/go-interact/interact"
)

const (
	red   color.Attribute = color.FgRed
	green                 = color.FgGreen
	// yellow                         = color.FgYellow
	// magenta                        = color.FgMagenta
	cyan = color.FgCyan
	// grey                           = color.FgWhite
	defaultFgColor = 38
)

//go:generate counterfeiter . Config

// Config is the UI configuration
type Config interface {
	// ColorEnabled enables or disabled color
	ColorEnabled() configv3.ColorSetting
	// Locale is the language to translate the output to
	Locale() string
}

//go:generate counterfeiter . TranslatableError

// TranslatableError it wraps the error interface adding a way to set the
// translation function on the error
type TranslatableError interface {
	// Returns back the untranslated error string
	Error() string
	Translate(func(string, ...interface{}) string) string
}

// UI is interface to interact with the user
type UI struct {
	// In is the input buffer
	In io.Reader
	// Out is the output buffer
	Out io.Writer
	// Err is the error buffer
	Err io.Writer

	colorEnabled configv3.ColorSetting
	translate    i18n.TranslateFunc
}

// NewUI will return a UI object where Out is set to STDOUT, In is set to
// STDIN, and Err is set to STDERR
func NewUI(c Config) (*UI, error) {
	translateFunc, err := GetTranslationFunc(c)
	if err != nil {
		return nil, err
	}

	return &UI{
		In:           os.Stdin,
		Out:          color.Output,
		Err:          os.Stderr,
		colorEnabled: c.ColorEnabled(),
		translate:    translateFunc,
	}, nil
}

// NewTestUI will return a UI object where Out, In, and Err are customizable,
// and colors are disabled
func NewTestUI(in io.Reader, out io.Writer, err io.Writer) *UI {
	return &UI{
		In:           in,
		Out:          out,
		Err:          err,
		colorEnabled: configv3.ColorDisabled,
		translate:    translationWrapper(i18n.IdentityTfunc()),
	}
}

// TranslateText passes the template through an internationalization function
// to translate it to a pre-configured language, and returns the template with
// templateValues substituted in. Only the first map in templateValues is used.
func (ui *UI) TranslateText(template string, templateValues ...map[string]interface{}) string {
	return ui.translate(template, getFirstSet(templateValues))
}

// DisplayOK outputs a bold green translated "OK" to UI.Out.
func (ui *UI) DisplayOK() {
	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText("OK"), green, true))
}

// DisplayNewline outputs a newline to UI.Out.
func (ui *UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayBoolPrompt outputs the prompt and waits for user input. It only
// allows for a boolean response. A default boolean response can be set with
// defaultResponse.
func (ui *UI) DisplayBoolPrompt(prompt string, defaultResponse bool) (bool, error) {
	response := defaultResponse
	interactivePrompt := interact.NewInteraction(fmt.Sprintf("%s%s", prompt, ui.addFlavor(">>", cyan, true)))
	interactivePrompt.Input = ui.In
	interactivePrompt.Output = ui.Out
	err := interactivePrompt.Resolve(&response)
	return response, err
}

// DisplayTable outputs a matrix of strings as a table to UI.Out. Prefix will
// be prepended to each row and padding adds the specified number of spaces
// between columns.
func (ui *UI) DisplayTable(prefix string, table [][]string, padding int) error {
	tw := tabwriter.NewWriter(ui.Out, 0, 1, padding, ' ', 0)
	for _, row := range table {
		fmt.Fprint(tw, prefix)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// DisplayText translates the template, substitutes in templateValues, and
// outputs the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayText(template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, templateValues...))
}

// DisplayPair translates the attribute, translates the template, substitutes
// templateValues into the template, and outputs the pair to ui.Out. Only the
// first map in templateValues is used.
func (ui *UI) DisplayPair(attribute string, template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Out, "%s: %s\n", ui.TranslateText(attribute), ui.TranslateText(template, templateValues...))
}

// DisplayHeader translates the header, bolds and adds the default color to the
// header, and outputs the result to ui.Out.
func (ui *UI) DisplayHeader(text string) {
	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText(text), defaultFgColor, true))
}

// DisplayTextWithFlavor translates the template, bolds and adds cyan color to
// templateValues, substitutes templateValues into the template, and outputs
// the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayTextWithFlavor(template string, templateValues ...map[string]interface{}) {
	firstTemplateValues := getFirstSet(templateValues)
	for key, value := range firstTemplateValues {
		firstTemplateValues[key] = ui.addFlavor(fmt.Sprint(value), cyan, true)
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, firstTemplateValues))
}

// DisplayWarning translates the warning, substitutes in templateValues, and
// outputs to ui.Err. Only the first map in templateValues is used.
func (ui *UI) DisplayWarning(template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(template, templateValues...))
}

// DisplayWarnings translates the warnings and outputs to ui.Err.
func (ui *UI) DisplayWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(warning))
	}
}

// DisplayError outputs the translated error message to ui.Err if the error
// satisfies TranslatableError, otherwise it outputs the original error message
// to ui.Err. It also outputs "FAILED" in bold red to ui.Out.
func (ui *UI) DisplayError(err error) {
	var errMsg string
	if translatableError, ok := err.(TranslatableError); ok {
		errMsg = translatableError.Translate(ui.translate)
	} else {
		errMsg = err.Error()
	}
	fmt.Fprintf(ui.Err, "%s\n", errMsg)
	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText("FAILED"), red, true))
}

// addFlavor adds the provided text color and bold style to the text.
func (ui *UI) addFlavor(text string, textColor color.Attribute, isBold bool) string {
	colorPrinter := color.New(textColor)

	switch ui.colorEnabled {
	case configv3.ColorEnabled:
		colorPrinter.EnableColor()
	case configv3.ColorDisabled:
		colorPrinter.DisableColor()
	}

	if isBold {
		colorPrinter = colorPrinter.Add(color.Bold)
	}

	printer := colorPrinter.SprintFunc()
	return printer(text)
}

// getFirstSet returns the first map if 1 or more maps are provided. Otherwise
// it returns the empty map.
func getFirstSet(list []map[string]interface{}) map[string]interface{} {
	if list == nil || len(list) == 0 {
		return map[string]interface{}{}
	}
	return list[0]
}
