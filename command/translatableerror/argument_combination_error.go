package translatableerror

import "strings"

// ArgumentCombinationError represent an error caused by using two command line
// arguments that cannot be used together.
type ArgumentCombinationError struct {
	Args []string
}

func (ArgumentCombinationError) DisplayUsage() {}

func (ArgumentCombinationError) Error() string {
	return "Incorrect Usage: The following arguments cannot be used together: {{.Args}}"
}

func (e ArgumentCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Args": strings.Join(e.Args, ", "),
	})
}
