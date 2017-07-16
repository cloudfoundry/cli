package translatableerror

type RequiredArgumentError struct {
	ArgumentName string
}

func (RequiredArgumentError) DisplayUsage() {}

func (RequiredArgumentError) Error() string {
	return "Incorrect Usage: the required argument `{{.ArgumentName}}` was not provided"
}

func (e RequiredArgumentError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName": e.ArgumentName,
	})
}
