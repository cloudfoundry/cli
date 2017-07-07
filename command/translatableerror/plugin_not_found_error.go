package translatableerror

type PluginNotFoundError struct {
	PluginName string
}

func (e PluginNotFoundError) Error() string {
	return "Plugin {{.PluginName}} does not exist."
}

func (e PluginNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PluginName": e.PluginName,
	})
}
