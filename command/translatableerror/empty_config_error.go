package translatableerror

type EmptyConfigError struct {
	FilePath string
}

func (EmptyConfigError) Error() string {
	return "Warning: Error read/writing config: unexpected end of JSON input for {{.FilePath}}"
}

func (e EmptyConfigError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"FilePath": e.FilePath,
	})
}
