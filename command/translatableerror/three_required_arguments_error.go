package translatableerror

type ThreeRequiredArgumentsError struct {
	ArgumentName1 string
	ArgumentName2 string
	ArgumentName3 string
}

func (ThreeRequiredArgumentsError) DisplayUsage() {}

func (ThreeRequiredArgumentsError) Error() string {
	return "Incorrect Usage: the required arguments `{{.ArgumentName1}}`, `{{.ArgumentName2}}`, and `{{.ArgumentName3}}` were not provided"
}

func (e ThreeRequiredArgumentsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName1": e.ArgumentName1,
		"ArgumentName2": e.ArgumentName2,
		"ArgumentName3": e.ArgumentName3,
	})
}
