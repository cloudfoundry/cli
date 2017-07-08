package shared

// PluginInstallationCancelled is used to ignore the scenario when the user
// responds with 'no' when prompted to install plugin and exit 0.
type PluginInstallationCancelled struct {
}

func (PluginInstallationCancelled) Error() string {
	return "Plugin installation cancelled"
}
