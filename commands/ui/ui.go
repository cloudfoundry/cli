// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
package ui

import (
	"os"
	"text/template"
)

// UI is the interface to STDOUT
type UI struct{}

// NewUI will return a UI object
func NewUI() UI {
	return UI{}
}

// DisplayText captures the input text in the Fake UI buffer so that this can
// be asserted against in tests. If multiple maps are passed in, the merge will
// give precedence to the latter maps.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	formattedTemplate := template.Must(template.New("Display Text").Parse(formattedString + "\n"))
	formattedTemplate.Execute(os.Stdout, ui.mergeMap(keys))
}

// DisplayTextWithKeyTranslations captures the input text in the Fake UI buffer
// so that this can be asserted against in tests. If multiple maps are passed
// in, the merge will give precedence to the latter maps. The list of
// keysToTranslate will then be translated prior to the string formatting.
func (ui UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	ui.DisplayText(formattedString, keys...)
}

// DisplayFlavorText captures the input text in the Fake UI buffer so that this
// can be asserted against in tests. If TTY is false, flavour text will not be
// outputted to STDOUT. If multiple maps are passed in, the merge will give
// precedence to the latter maps.
func (ui UI) DisplayFlavorText(formattedString string, keys ...map[string]interface{}) {
	ui.DisplayText(formattedString, keys...)
}

// DisplayFlavorTextWithKeyTranslations captures the input text in the Fake UI
// buffer so that this can be asserted against in tests. If TTY is false,
// flavour text will not be outputted to STDOUT. If multiple maps are passed
// in, the merge will give precedence to the latter maps. The list of
// keysToTranslate will then be translated prior to the string formatting.
func (ui UI) DisplayFlavorTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	ui.DisplayText(formattedString, keys...)
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
