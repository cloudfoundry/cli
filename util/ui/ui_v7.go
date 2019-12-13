// +build V7

package ui

import (
	"fmt"
)

// DisplayWarning translates the warning, substitutes in templateValues, and
// outputs to ui.Err. Only the first map in templateValues is used.
// This command has one fewer newline than DisplayWarning. Use it before an OK message in V7.
func (ui *UI) DisplayWarning(template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(template, templateValues...))
}

// Translates warnings and outputs them to ui.Err.
// Prints each warning with a trailing newline.
func (ui *UI) DisplayWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(warning))
	}
}
