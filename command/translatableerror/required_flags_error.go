package translatableerror

// RequiredFlagsError represent an error caused by using a command line
// argument that requires another flags to be used.
type RequiredFlagsError struct {
	Arg1 string
	Arg2 string
}

func (RequiredFlagsError) DisplayUsage() {}

func (RequiredFlagsError) Error() string {
	return "Incorrect Usage: '{{.Arg1}}' and '{{.Arg2}}' must be used together."
}

func (e RequiredFlagsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg1": e.Arg1,
		"Arg2": e.Arg2,
	})
}
