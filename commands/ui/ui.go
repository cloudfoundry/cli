// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
package ui

import (
	"fmt"
	"io"
	"text/template"

	"github.com/fatih/color"

	"code.cloudfoundry.org/cli/utils/config"
	"github.com/nicksnyder/go-i18n/i18n"
)

const (
	red            color.Attribute = color.FgRed
	green                          = color.FgGreen
	yellow                         = color.FgYellow
	magenta                        = color.FgMagenta
	cyan                           = color.FgCyan
	grey                           = color.FgWhite
	defaultFgColor                 = 38
)

//go:generate counterfeiter . Config

// Config is the UI configuration
type Config interface {
	// ColorEnabled enables or disabled color
	ColorEnabled() config.ColorSetting

	// Locale is the language to translate the output to
	Locale() string
}

// UI is interface to interact with the user
type UI struct {
	// Out is the output buffer
	Out io.Writer

	colorEnabled config.ColorSetting

	translate i18n.TranslateFunc
}

// NewUI will return a UI object where Out is set to STDOUT
func NewUI(c Config) (UI, error) {
	translateFunc, err := GetTranslationFunc(c)
	if err != nil {
		return UI{}, err
	}

	return UI{
		Out:          color.Output,
		colorEnabled: c.ColorEnabled(),
		translate:    translateFunc,
	}, nil
}

// NewTestUI will return a UI object where Out is customizable
func NewTestUI(out io.Writer) UI {
	return UI{
		Out:          out,
		colorEnabled: config.ColorDisbled,
		translate:    i18n.TranslateFunc(func(s string, _ ...interface{}) string { return s }),
	}
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. The maps are merged in a way that the last
// one takes precidence over the first. Prior to outputting the
// formattedString, it is run through the an internationalization function to
// translate it to a pre-cofigured langauge.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	translatedFormatString := ui.translate(formattedString, mergedMap)
	formattedTemplate := template.Must(template.New("Display Text").Parse(translatedFormatString))
	formattedTemplate.Execute(ui.Out, mergedMap)
	ui.DisplayNewline()
}

// DisplayTextWithKeyTranslations merges keys together (similar to
// DisplayText), translates the keys listed in keysToTranslate, and then passes
// these values to DisplayText.
func (ui UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	for _, key := range keysToTranslate {
		mergedMap[key] = ui.translate(mergedMap[key].(string))
	}
	ui.DisplayText(formattedString, mergedMap)
}

// DisplayNewline outputs a newline.
func (ui UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayHelpHeader translates and then bolds the help header.
func (ui UI) DisplayHelpHeader(text string) {
	fmt.Fprintf(ui.Out, ui.colorize(ui.translate(text), defaultFgColor, true))
	ui.DisplayNewline()
}

func (ui UI) mergeMap(maps []map[string]interface{}) map[string]interface{} {
	if len(maps) == 1 {
		return maps[0]
	}

	main := map[string]interface{}{}

	for _, minor := range maps {
		for key, value := range minor {
			main[key] = value
		}
	}

	return main
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
