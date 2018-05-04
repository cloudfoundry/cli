package command

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func MinimumAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	comparison, err := compareSemVer(current, minimum)
	if err != nil {
		return err
	}

	if comparison == -1 {
		return translatableerror.MinimumAPIVersionNotMetError{
			Command:        command,
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
