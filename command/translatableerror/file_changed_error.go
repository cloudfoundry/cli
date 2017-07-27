package translatableerror

type FileChangedError struct {
	Filename string
}

func (e FileChangedError) Error() string {
	return "Aborting push: File {{.Filename}} has been modified since the start of push. Validate the correct state of the file and try again."
}

func (e FileChangedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Filename": e.Filename,
	})
}
