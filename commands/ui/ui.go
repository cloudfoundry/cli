// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
package ui

import (
	"fmt"
	"io"
	"text/template"

	"github.com/fatih/color"

	"code.cloudfoundry.org/cli/utils/config"
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

type Config interface {
	ColorEnabled() config.ColorSetting
}

// UI is interface to interact with the user
type UI struct {
	// Out is the output buffer.
	Out io.Writer

	colorEnabled config.ColorSetting
}

// NewUI will return a UI object where Out is set to STDOUT
func NewUI(c Config) UI {
	return UI{
		Out:          color.Output,
		colorEnabled: c.ColorEnabled(),
	}
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. The maps are merged in a way that the last
// one takes precidence over the first.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	formattedTemplate := template.Must(template.New("Display Text").Parse(formattedString))
	formattedTemplate.Execute(ui.Out, ui.mergeMap(keys))
	ui.DisplayNewline()
}

// DisplayTextWithKeyTranslations captures the input text in the Fake UI buffer
// so that this can be asserted against in tests. If multiple maps are passed
// in, the merge will give precedence to the latter maps. The list of
// keysToTranslate will then be translated prior to the string formatting.
func (ui UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	ui.DisplayText(formattedString, keys...)
}

// DisplayNewline outputs a newline.
func (ui UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayHelpHeader outputs a bolded help header
func (ui UI) DisplayHelpHeader(text string) {
	ui.DisplayText(ui.colorize(text, defaultFgColor, true))
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
