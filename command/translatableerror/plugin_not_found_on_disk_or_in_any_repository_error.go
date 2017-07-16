package translatableerror

type PluginNotFoundOnDiskOrInAnyRepositoryError struct {
	PluginName string
	BinaryName string
}

func (e PluginNotFoundOnDiskOrInAnyRepositoryError) Error() string {
	return "Plugin {{.PluginName}} not found on disk or in any registered repo.\nUse '{{.BinaryName}} repo-plugins' to list plugins available in the repos."
}

func (e PluginNotFoundOnDiskOrInAnyRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName": e.PluginName,
		"BinaryName": e.BinaryName,
	})
}
