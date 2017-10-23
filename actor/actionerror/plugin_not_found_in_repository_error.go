package actionerror

import "fmt"

// PluginNotFoundInRepositoryError is an error returned when a plugin is not
// found.
type PluginNotFoundInRepositoryError struct {
	PluginName     string
	RepositoryName string
}

// Error outputs the plugin not found in repository error message.
func (e PluginNotFoundInRepositoryError) Error() string {
	return fmt.Sprintf("Plugin %s not found in repository %s", e.PluginName, e.RepositoryName)
}
