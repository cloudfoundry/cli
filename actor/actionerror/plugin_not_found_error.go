package actionerror

import "fmt"

// PluginNotFoundError is an error returned when a plugin is not found.
type PluginNotFoundError struct {
	PluginName string
}

// Error outputs a plugin not found error message.
func (e PluginNotFoundError) Error() string {
	return fmt.Sprintf("Plugin name %s does not exist.", e.PluginName)
}
