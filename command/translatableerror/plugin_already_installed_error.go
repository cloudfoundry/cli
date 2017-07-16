package translatableerror

// PluginAlreadyInstalledError is returned when the plugin has the same name as
// an installed plugin.
type PluginAlreadyInstalledError struct {
	BinaryName string
	Name       string
	Version    string
}

func (PluginAlreadyInstalledError) Error() string {
	return "Plugin {{.Name}} {{.Version}} could not be installed. A plugin with that name is already installed.\nTIP: Use '{{.BinaryName}} install-plugin -f' to force a reinstall."
}

func (e PluginAlreadyInstalledError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BinaryName": e.BinaryName,
		"Name":       e.Name,
		"Version":    e.Version,
	})
}
