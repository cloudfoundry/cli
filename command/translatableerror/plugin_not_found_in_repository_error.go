package translatableerror

type PluginNotFoundInRepositoryError struct {
	BinaryName     string
	PluginName     string
	RepositoryName string
}

func (e PluginNotFoundInRepositoryError) Error() string {
	return "Plugin {{.PluginName}} not found in repository {{.RepositoryName}}.\nUse '{{.BinaryName}} repo-plugins -r {{.RepositoryName}}' to list plugins available in the repo."
}

func (e PluginNotFoundInRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName":     e.PluginName,
		"RepositoryName": e.RepositoryName,
		"BinaryName":     e.BinaryName,
	})
}
