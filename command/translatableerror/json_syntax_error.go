package translatableerror

type JSONSyntaxError struct {
	Err error
}

func (e JSONSyntaxError) Error() string {
	return "Invalid JSON content from server: {{.Err}}"
}

func (e JSONSyntaxError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err.Error(),
	})
}
