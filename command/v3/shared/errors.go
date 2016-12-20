package shared

type RunTaskError struct {
	Message string
}

func (e RunTaskError) Error() string {
	return "Error running task: {{.CloudControllerMessage}}"
}

func (e RunTaskError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CloudControllerMessage": e.Message,
	})
}

type ClientTargetError struct {
	Message string
}

func (e ClientTargetError) Error() string {
	return "{{.Message}}\nNote that this command requires CF API version 3.0.0+."
}

func (e ClientTargetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
