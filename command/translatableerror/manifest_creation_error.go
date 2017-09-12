package translatableerror

type ManifestCreationError struct {
	Err error
}

func (ManifestCreationError) Error() string {
	return "Error creating manifest file: {{.Error}}"
}

func (e ManifestCreationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}
