package translatableerror

import "fmt"

type NotLoggedInError struct {
	BinaryName string
}

func (NotLoggedInError) Error() string {
	return "Not logged in. Use '{{.CFLoginCommand}}' or '{{.CFLoginCommandSSO}}' to log in."
}

func (e NotLoggedInError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CFLoginCommand":    fmt.Sprintf("%s login", e.BinaryName),
		"CFLoginCommandSSO": fmt.Sprintf("%s login --sso", e.BinaryName),
	})
}
