// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
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

	"github.com/fatih/color"

	"code.cloudfoundry.org/cli/utils/config"
	"github.com/nicksnyder/go-i18n/i18n"
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
	ColorEnabled() config.ColorSetting

	// Locale is the language to translate the output to
	Locale() string
}

//go:generate counterfeiter . TranslatableError

// TranslatableError it wraps the error interface adding a way to set the
// translation function on the error
type TranslatableError interface {
	// Returns back the untranslated error string
	Error() string
	SetTranslation(i18n.TranslateFunc) error
}

// UI is interface to interact with the user
type UI struct {
	// Out is the output buffer
	Out io.Writer

	// Err is the error buffer
	Err io.Writer

	colorEnabled config.ColorSetting

	translate i18n.TranslateFunc
}

// NewUI will return a UI object where Out is set to STDOUT and Err is set to
// STDERR
func NewUI(c Config) (UI, error) {
	translateFunc, err := GetTranslationFunc(c)
	if err != nil {
		return UI{}, err
	}

	return UI{
		Out:          color.Output,
		Err:          os.Stderr,
		colorEnabled: c.ColorEnabled(),
		translate:    translateFunc,
	}, nil
}

// NewTestUI will return a UI object where Out and Err are customizable, and
// colors are disabled
func NewTestUI(out io.Writer, err io.Writer) UI {
	return UI{
		Out:          out,
		Err:          err,
		colorEnabled: config.ColorDisbled,
		translate:    translationWrapper(i18n.IdentityTfunc()),
	}
}

// DisplayTable presents a two dimensional array of strings as a table to UI.Out
func (ui UI) DisplayTable(prefix string, table [][]string) {
	tw := tabwriter.NewWriter(ui.Out, 0, 1, 4, ' ', 0)

	for _, row := range table {
		fmt.Fprint(tw, prefix)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	tw.Flush()
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. Prior to outputting the formattedString, it
// is run through an internationalization function to translate it to a
// pre-configured language. Only the first map in keys is used.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(formattedString, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Out, "%s\n", translatedValue)
}

// DisplayTextWithKeyTranslations translates the keys listed in
// keysToTranslate, and then passes these values to DisplayText. Only the first
// map in keys is used.
func (ui UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	templateValues := ui.templateValuesFromKeys(keys)
	for _, key := range keysToTranslate {
		templateValues[key] = ui.translate(templateValues[key].(string))
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.translate(formattedString, templateValues))
}

// DisplayNewline outputs a newline to UI.Out.
func (ui UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayPair outputs the "attribute: formattedString" pair to UI.Out. keys
// are applied to the translation of formattedString, while attribute is
// translated directly.
func (ui UI) DisplayPair(attribute string, formattedString string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(formattedString, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Out, "%s: %s\n", ui.translate(attribute), translatedValue)
}

// DisplayHelpHeader translates and then bolds the help header. Sends output to
// UI.Out.
func (ui UI) DisplayHelpHeader(text string) {
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(ui.translate(text), defaultFgColor, true))
}

// DisplayHeaderFlavorText outputs the translated text, with cyan color keys,
// to UI.Out.
func (ui UI) DisplayHeaderFlavorText(formattedString string, keys ...map[string]interface{}) {
	templateValues := ui.templateValuesFromKeys(keys)
	for key, value := range templateValues {
		templateValues[key] = ui.colorize(fmt.Sprint(value), cyan, true)
	}

	translatedValue := ui.translate(formattedString, templateValues)
	fmt.Fprintf(ui.Out, "%s\n", translatedValue)
}

// DisplayOK outputs a green translated "OK" message to UI.Out.
func (ui UI) DisplayOK() {
	translatedFormatString := ui.translate("OK", nil)
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(translatedFormatString, green, true))
}

// DisplayErrorMessage combines the err template with the key maps and then
// outputs it to the UI.Err file. It will then output a red translated "FAILED"
// to UI.Out. Prior to outputting the err, it is run through an
// internationalization function to translate it to a pre-configured language.
func (ui UI) DisplayErrorMessage(err string, keys ...map[string]interface{}) {
	translatedValue := ui.translate(err, ui.templateValuesFromKeys(keys))
	fmt.Fprintf(ui.Err, "%s\n", translatedValue)

	translatedFormatString := ui.translate("FAILED", nil)
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(translatedFormatString, red, true))
}

// DisplayError outputs the error to UI.Err and outputs a red translated
// "FAILED" to UI.Out.
func (ui UI) DisplayError(originalErr TranslatableError) {
	err := originalErr.SetTranslation(ui.translate)
	fmt.Fprintf(ui.Err, "%s\n", err.Error())

	translatedFormatString := ui.translate("FAILED", nil)
	fmt.Fprintf(ui.Out, "%s\n", ui.colorize(translatedFormatString, red, true))
}

func (ui UI) templateValuesFromKeys(keys []map[string]interface{}) map[string]interface{} {
	if len(keys) > 0 {
		return keys[0]
	}
	return map[string]interface{}{}
}

func (ui UI) colorize(message string, textColor color.Attribute, bold bool) string {
	colorPrinter := color.New(textColor)
	switch ui.colorEnabled {
	case config.ColorEnabled:
		colorPrinter.EnableColor()
	case config.ColorDisbled:
		colorPrinter.DisableColor()
	}

	if bold {
		colorPrinter = colorPrinter.Add(color.Bold)
	}
	f := colorPrinter.SprintFunc()
	return f(message)
}
