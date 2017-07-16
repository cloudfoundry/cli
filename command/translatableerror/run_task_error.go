package translatableerror

type RunTaskError struct {
	Message string
}

func (RunTaskError) Error() string {
	return "Error running task: {{.CloudControllerMessage}}"
}

func (e RunTaskError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CloudControllerMessage": e.Message,
	})
}
