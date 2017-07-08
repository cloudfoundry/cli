package translatableerror

type APINotFoundError struct {
	URL string
}

func (APINotFoundError) Error() string {
	return "API endpoint not found at '{{.URL}}'"
}

func (e APINotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.URL,
	})
}
