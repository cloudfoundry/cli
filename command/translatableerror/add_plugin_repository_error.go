package translatableerror

type AddPluginRepositoryError struct {
	Name    string
	URL     string
	Message string
}

func (AddPluginRepositoryError) Error() string {
	return "Could not add repository '{{.RepositoryName}}' from {{.RepositoryURL}}: {{.Message}}"
}

func (e AddPluginRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.Name,
		"RepositoryURL":  e.URL,
		"Message":        e.Message,
	})
}
