package translatableerror

type FetchingPluginInfoFromRepositoriesError struct {
	Message        string
	RepositoryName string
}

func (FetchingPluginInfoFromRepositoriesError) Error() string {
	return "Plugin list download failed; repository {{.RepositoryName}} returned {{.ErrorMessage}}."
}

func (e FetchingPluginInfoFromRepositoriesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.RepositoryName,
		"ErrorMessage":   e.Message,
	})
}
