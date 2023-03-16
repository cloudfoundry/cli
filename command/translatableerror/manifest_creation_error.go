package translatableerror

type FileCreationError struct {
	Err error
}

func (FileCreationError) Error() string {
	return "Error creating file: {{.Error}}"
}

func (e FileCreationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}
