package translatableerror

type V3APIDoesNotExistError struct {
	Message string
}

func (V3APIDoesNotExistError) Error() string {
	return "{{.Message}}\nThis command requires CF API version 3.0.0 or higher."
}

func (e V3APIDoesNotExistError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
