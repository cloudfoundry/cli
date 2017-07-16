package translatableerror

// PluginBinaryUninstallError is returned when running the plugin's uninstall
// hook fails.
type PluginBinaryUninstallError struct {
	Err error
}

func (e PluginBinaryUninstallError) Error() string {
	return "The plugin's uninstall method returned an unexpected error.\nThe plugin uninstall will proceed. Contact the plugin author if you need help.\n{{.Err}}"
}

func (e PluginBinaryUninstallError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Err": e.Err,
	})
}
