package translatableerror

type IncorrectUsageError struct {
	Message string
}

func (IncorrectUsageError) DisplayUsage() {}

func (IncorrectUsageError) Error() string {
	return "Incorrect Usage: {{.Message}}"
}

func (e IncorrectUsageError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
