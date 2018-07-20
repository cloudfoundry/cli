package translatableerror

import (
	"fmt"
	"strings"
)

type MultiError struct {
	Messages []string
}

func (MultiError) Error() string {
	return "Multiple errors occurred:\n{{.Errors}}"
}

func (e MultiError) Translate(translate func(string, ...interface{}) string) string {
	var formattedErrs []string
	for _, err := range e.Messages {
		formattedErrs = append(formattedErrs, fmt.Sprintf("- %s", err))
	}
	return translate(e.Error(), map[string]interface{}{
		"Errors": strings.Join(formattedErrs, "\n"),
	})
}
