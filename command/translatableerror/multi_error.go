package translatableerror

import "strings"

type MultiError struct {
	Messages []string
}

func (MultiError) Error() string {
	return "Multiple errors returned:\n{{.Errors}}"
}

func (e MultiError) Translate(translate func(string, ...interface{}) string) string {

	return translate(e.Error(), map[string]interface{}{
		"Errors": strings.Join(e.Messages, "\n"),
	})
}
