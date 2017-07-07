package translatableerror

type V3APIDoesNotExistError struct {
	Message string
}

func (_ V3APIDoesNotExistError) Error() string {
	return "{{.Message}}\nNote that this command requires CF API version 3.0.0+."
}

func (e V3APIDoesNotExistError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
