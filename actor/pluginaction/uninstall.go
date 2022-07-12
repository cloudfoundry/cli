package pluginaction

import (
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PluginUninstaller

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

		removed := false
		for i := 0; i < 50; i++ {
			err = os.Remove(plugin.Location)
			if err == nil {
				removed = true
				break
			}

			if os.IsNotExist(err) {
				removed = true
				break
			}

			if !os.IsNotExist(err) {
				if _, isPathError := err.(*os.PathError); isPathError {
					time.Sleep(50 * time.Millisecond)
				} else {
					return err
				}
			}
		}

		if !removed {
			binaryErr = actionerror.PluginBinaryRemoveFailedError{
				Err: err,
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
