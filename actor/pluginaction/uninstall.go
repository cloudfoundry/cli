package pluginaction

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// PluginNotFoundError is an error returned when a plugin is not found.
type PluginNotFoundError struct {
	PluginName string
}

// Error outputs a plugin not found error message.
func (e PluginNotFoundError) Error() string {
	return fmt.Sprintf("Plugin name %s does not exist.", e.PluginName)
}

//go:generate counterfeiter . PluginUninstaller

type PluginUninstaller interface {
	Run(pluginPath string, command string) error
}

func (actor Actor) UninstallPlugin(uninstaller PluginUninstaller, name string) error {
	plugin, exist := actor.config.GetPlugin(name)
	if !exist {
		return PluginNotFoundError{PluginName: name}
	}

	var binaryErr error

	if actor.FileExists(plugin.Location) {
		err := uninstaller.Run(plugin.Location, "CLI-MESSAGE-UNINSTALL")
		if err != nil {
			if _, isExitError := err.(*exec.ExitError); isExitError {
				binaryErr = err
			} else {
				return err
			}
		}

		// No test for sleeping for 500 ms for parity with pre-refactored behavior.
		time.Sleep(500 * time.Millisecond)

		err = os.Remove(plugin.Location)
		if err != nil && !os.IsNotExist(err) {
			if _, isPathError := err.(*os.PathError); isPathError {
				binaryErr = err
			} else {
				return err
			}
		}
	}

	actor.config.RemovePlugin(name)
	err := actor.config.WritePluginConfig()
	if err != nil {
		return err
	}

	return binaryErr
}
