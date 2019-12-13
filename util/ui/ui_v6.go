// +build !V7

package ui

import (
	"fmt"
)

// DisplayWarning translates the warning, substitutes in templateValues, and
// outputs to ui.Err. Only the first map in templateValues is used.
func (ui *UI) DisplayWarning(template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Err, "%s\n\n", ui.TranslateText(template, templateValues...))
}

// Translates warnings and outputs them to ui.Err.
// Prints each warning with a trailing newline.
// Prints the final warning with two trailing newlines.
func (ui *UI) DisplayWarnings(warnings []string) {

	for _, warning := range warnings {
		fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(warning))
	}
	if len(warnings) > 0 {
		fmt.Fprintln(ui.Err)
	}
}
