// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
package ui

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

// UI is interface to interact with the user
type UI struct {
	// Out is the output buffer.
	Out io.WriteCloser
}

// NewUI will return a UI object where Out is set to STDOUT
func NewUI() UI {
	InitColorSupport()

	return UI{
		Out: os.Stdout,
	}
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. The maps are merged in a way that the last
// one takes precidence over the first.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	formattedTemplate := template.Must(template.New("Display Text").Parse(formattedString + "\n"))
	formattedTemplate.Execute(ui.Out, ui.mergeMap(keys))
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

// DisplayHelpHeader outputs a help header
func (ui UI) DisplayHelpHeader(text string) {
	ui.DisplayText(colorize(text, defaultFgColor, true))
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
