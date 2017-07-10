package translatableerror

type ParseArgumentError struct {
	ArgumentName string
	ExpectedType string
}

func (ParseArgumentError) DisplayUsage() {}

func (ParseArgumentError) Error() string {
	return "Incorrect usage: Value for {{.ArgumentName}} must be {{.ExpectedType}}"
}

func (e ParseArgumentError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName": e.ArgumentName,
		"ExpectedType": e.ExpectedType,
	})
}
