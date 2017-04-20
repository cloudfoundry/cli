package pluginaction

import (
	"fmt"
	"os"
	"time"
)

// PluginNotFoundError is an error returned when a plugin is not found.
type PluginNotFoundError struct {
	Name string
}

// Error outputs a plugin not found error message.
func (e PluginNotFoundError) Error() string {
	return fmt.Sprintf("Plugin name %s does not exist.", e.Name)
}

//go:generate counterfeiter . PluginUninstaller

type PluginUninstaller interface {
	Uninstall(pluginPath string) error
}

func (actor Actor) UninstallPlugin(uninstaller PluginUninstaller, name string) error {
	plugin, exist := actor.config.GetPlugin(name)
	if !exist {
		return PluginNotFoundError{Name: name}
	}

	err := uninstaller.Uninstall(plugin.Location)
	if err != nil {
		return err
	}

	// No test for sleeping for 500 ms for parity with pre-refactored behavior.
	time.Sleep(500 * time.Millisecond)

	err = os.Remove(plugin.Location)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	actor.config.RemovePlugin(name)
	return actor.config.WritePluginConfig()
}
