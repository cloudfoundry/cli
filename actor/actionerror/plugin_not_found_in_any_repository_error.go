package actionerror

import "fmt"

// PluginNotFoundInAnyRepositoryError is an error returned when a plugin cannot
// be found in any repositories.
type PluginNotFoundInAnyRepositoryError struct {
	PluginName string
}

// Error outputs that the plugin cannot be found in any repositories.
func (e PluginNotFoundInAnyRepositoryError) Error() string {
	return fmt.Sprintf("Plugin %s not found in any registered repo", e.PluginName)
}
