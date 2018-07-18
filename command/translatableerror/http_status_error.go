package translatableerror

type HTTPStatusError struct {
	Status string
}

func (e HTTPStatusError) Error() string {
	return "Download attempt failed; server returned {{.Status}}\nUnable to install; buildpack is not available from the given URL."
}

func (e HTTPStatusError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Status": e.Status,
	})
}
