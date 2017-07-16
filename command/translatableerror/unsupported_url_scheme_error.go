package translatableerror

type UnsupportedURLSchemeError struct {
	UnsupportedURL string
}

func (e UnsupportedURLSchemeError) Error() string {
	return "This command does not support the URL scheme in {{.UnsupportedURL}}."
}

func (e UnsupportedURLSchemeError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"UnsupportedURL": e.UnsupportedURL,
	})
}
