package translatableerror

type TooManyArgumentsError struct {
	ExtraArgument string
}

func (TooManyArgumentsError) Error() string {
	return `Incorrect Usage: unexpected argument "{{.ExtraArgument}}"`
}

func (e TooManyArgumentsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ExtraArgument": e.ExtraArgument,
	})
}

func (TooManyArgumentsError) DisplayUsage() {}
