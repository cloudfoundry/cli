package translatableerror

type EmptyArchiveError struct {
	Path string
}

func (e EmptyArchiveError) Error() string {
	return "No app files found in '{{.Path}}'"
}

func (e EmptyArchiveError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Path": e.Path,
	})
}
