package version

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"github.com/blang/semver"
)

func MinimumAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	if current == DefaultVersion || minimum == "" {
		return nil
	}

	currentSemvar, err := semver.Make(current)
	if err != nil {
		return err
	}

	minimumSemvar, err := semver.Make(minimum)
	if err != nil {
		return err
	}

	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	if currentSemvar.Compare(minimumSemvar) == -1 {
		return translatableerror.MinimumAPIVersionNotMetError{
			Command:        command,
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
