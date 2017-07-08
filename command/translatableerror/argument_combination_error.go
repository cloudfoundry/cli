package translatableerror

// ArgumentCombinationError represent an error caused by using two command line
// arguments that cannot be used together.
type ArgumentCombinationError struct {
	Arg1 string
	Arg2 string
}

func (ArgumentCombinationError) Error() string {
	return "Incorrect Usage: '{{.Arg1}}' and '{{.Arg2}}' cannot be used together."
}

func (e ArgumentCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg1": e.Arg1,
		"Arg2": e.Arg2,
	})
}
