package translatableerror

type StagingFailedError struct {
	Message string
}

func (StagingFailedError) Error() string {
	return "Error staging application: {{.Message}}"
}

func (e StagingFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
