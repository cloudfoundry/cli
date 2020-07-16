package translatableerror

type RequestCreationError struct {
	Err error
}

func (RequestCreationError) Error() string {
	return "Error creating request: {{.Error}}"
}

func (e RequestCreationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}
