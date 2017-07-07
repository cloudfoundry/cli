package translatableerror

type EmptyDirectoryError struct {
	Path string
}

func (e EmptyDirectoryError) Error() string {
	return "No app files found in '{{.Path}}'"
}

func (e EmptyDirectoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Path": e.Path,
	})
}
