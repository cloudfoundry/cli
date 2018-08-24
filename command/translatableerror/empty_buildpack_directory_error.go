package translatableerror

type EmptyBuildpackDirectoryError struct {
	Path string
}

func (EmptyBuildpackDirectoryError) Error() string {
	return "The specified path '{{.Path}}' cannot be an empty directory."
}

func (e EmptyBuildpackDirectoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Path": e.Path,
	})
}
