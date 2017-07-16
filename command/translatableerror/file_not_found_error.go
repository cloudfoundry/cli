package translatableerror

// FileNotFoundError is returned when a local plugin binary is not found during
// installation.
type FileNotFoundError struct {
	Path string
}

func (FileNotFoundError) Error() string {
	return "File not found locally, make sure the file exists at given path {{.FilePath}}"
}

func (e FileNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"FilePath": e.Path,
	})
}
