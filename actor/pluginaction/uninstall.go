package pluginaction

import (
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
)

//go:generate counterfeiter . PluginUninstaller

type PluginUninstaller interface {
	Run(pluginPath string, command string) error
}

func (actor Actor) UninstallPlugin(uninstaller PluginUninstaller, name string) error {
	plugin, exist := actor.config.GetPlugin(name)
	if !exist {
		return actionerror.PluginNotFoundError{PluginName: name}
	}

	var binaryErr error

	if actor.FileExists(plugin.Location) {
		err := uninstaller.Run(plugin.Location, "CLI-MESSAGE-UNINSTALL")
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				binaryErr = actionerror.PluginExecuteError{
					Err: err,
				}
			case *os.PathError:
				binaryErr = actionerror.PluginExecuteError{
					Err: err,
				}
			default:
				return err
			}
		}

		// No test for sleeping for 500 ms for parity with pre-refactored behavior.
		time.Sleep(500 * time.Millisecond)

		err = os.Remove(plugin.Location)
		if err != nil && !os.IsNotExist(err) {
			if _, isPathError := err.(*os.PathError); isPathError {
				binaryErr = actionerror.PluginBinaryRemoveFailedError{
					Err: err,
				}
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
