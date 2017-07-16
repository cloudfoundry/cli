package translatableerror

// PluginBinaryRemoveFailedError is returned when the removal of a plugin
// binary fails.
type PluginBinaryRemoveFailedError struct {
	Err error
}

func (e PluginBinaryRemoveFailedError) Error() string {
	return "The plugin has been uninstalled but removing the plugin binary failed.\nRemove it manually or subsequent installations of the plugin may fail\n{{.Err}}"
}

func (e PluginBinaryRemoveFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err,
	})
}
