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

	"code.cloudfoundry.org/cli/utils/configv3"

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

	translate i18n.TranslateFunc
}

// NewUI will return a UI object where Out is set to STDOUT, In is set to STDIN,
// and Err is set to STDERR
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

// NewTestUI will return a UI object where Out, In, and Err are customizable, and
// colors are disabled
func NewTestUI(in io.Reader, out io.Writer, err io.Writer) *UI {
	return &UI{
		In:           in,
		Out:          out,
		Err:          err,
		colorEnabled: configv3.ColorDisabled,
		translate:    translationWrapper(i18n.IdentityTfunc()),
	}
}

// DisplayTable presents a two dimensional array of strings as a table to UI.Out
func (ui *UI) DisplayTable(prefix string, table [][]string) error {
	tw := tabwriter.NewWriter(ui.Out, 0, 1, 4, ' ', 0)

	for _, row := range table {
		fmt.Fprint(tw, prefix)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	return tw.Flush()
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. Prior to outputting the formattedString, it
// is run through an internationalization function to translate it to a
// pre-configured language. Only the first map in keys is used.
func (ui *UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(formattedString, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Out, "%s\n", translatedValue)
}

// DisplayTextWithKeyTranslations translates the keys listed in
// keysToTranslate, and then passes these values to DisplayText. Only the first
// map in keys is used.
func (ui *UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	templateValues := ui.templateValuesFromKeys(keys)
	for _, key := range keysToTranslate {
		templateValues[key] = ui.translate(templateValues[key].(string))
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.translate(formattedString, templateValues))
}

// DisplayNewline outputs a newline to UI.Out.
func (ui *UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayPair outputs the "attribute: formattedString" pair to UI.Out. keys
// are applied to the translation of formattedString, while attribute is
// translated directly.
func (ui *UI) DisplayPair(attribute string, formattedString string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(formattedString, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Out, "%s: %s\n", ui.translate(attribute), translatedValue)
}

// DisplayBoolPrompt outputs the prompt and waits for user input. It only
// allows for a boolean response. A default boolean response can be set with
// defaultResponse.
func (ui *UI) DisplayBoolPrompt(prompt string, defaultResponse bool) (bool, error) {
	response := defaultResponse
	fullPrompt := fmt.Sprintf("%s%s", prompt, ui.colorize(">>", cyan, true))
	interactivePrompt := interact.NewInteraction(fullPrompt)
	interactivePrompt.Input = ui.In
	interactivePrompt.Output = ui.Out
	err := interactivePrompt.Resolve(&response)
	return response, err
}

// DisplayHelpHeader translates and then bolds the help header. Sends output to
// UI.Out.
func (ui *UI) DisplayHelpHeader(text string) {
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(ui.translate(text), defaultFgColor, true))
}

// DisplayHeaderFlavorText outputs the translated text, with cyan color keys,
// to UI.Out.
func (ui *UI) DisplayHeaderFlavorText(formattedString string, keys ...map[string]interface{}) {
	templateValues := ui.templateValuesFromKeys(keys)
	for key, value := range templateValues {
		templateValues[key] = ui.colorize(fmt.Sprint(value), cyan, true)
	}

	translatedValue := ui.translate(formattedString, templateValues)
	fmt.Fprintf(ui.Out, "%s\n", translatedValue)
}

// DisplayOK outputs a green translated "OK" message to UI.Out.
func (ui *UI) DisplayOK() {
	translatedFormatString := ui.translate("OK", nil)
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(translatedFormatString, green, true))
}

// DisplayError outputs the error to UI.Err and outputs a red translated
// "FAILED" to UI.Out.
func (ui *UI) DisplayError(err error) {
	if translatableError, ok := err.(TranslatableError); ok {
		fmt.Fprintf(ui.Err, "%s\n", translatableError.Translate(ui.translate))
	} else {
		fmt.Fprintf(ui.Err, "%s\n", err.Error())
	}

	translatedFormatString := ui.translate("FAILED", nil)
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(translatedFormatString, red, true))
}

// DisplayWarning applies translation to formattedString and displays the
// translated warning to UI.Err.
func (ui *UI) DisplayWarning(formattedString string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(formattedString, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Err, "%s\n", translatedValue)
}

// DisplayWarnings translates and displays the warnings.
func (ui *UI) DisplayWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintf(ui.Err, "%s\n", ui.translate(warning, nil))
	}
}

func (ui *UI) templateValuesFromKeys(keys []map[string]interface{}) map[string]interface{} {
	if len(keys) > 0 {
		return keys[0]
	}
	return map[string]interface{}{}
}

func (ui *UI) colorize(message string, textColor color.Attribute, bold bool) string {
	colorPrinter := color.New(textColor)
	switch ui.colorEnabled {
	case configv3.ColorEnabled:
		colorPrinter.EnableColor()
	case configv3.ColorDisabled:
		colorPrinter.DisableColor()
	}

	if bold {
		colorPrinter = colorPrinter.Add(color.Bold)
	}
	f := colorPrinter.SprintFunc()
	return f(message)
}
