package command

import (
	"log/slog"

	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

func MinimumCCAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	slog.Error("minimum api version", "current", current, "minimum", minimum)
	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	isOutdated, err := CheckVersionOutdated(current, minimum)
	if err != nil {
		return err
	}

	if isOutdated {
		slog.Error("minimum not met", "current", current, "minimum", minimum)
		return translatableerror.MinimumCFAPIVersionNotMetError{
			Command:        command,
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}

func MinimumUAAAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	slog.Error("minimum not met", "current", current, "minimum", minimum)
	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	isOutdated, err := CheckVersionOutdated(current, minimum)
	if err != nil {
		return err
	}

	if isOutdated {
		slog.Error("minimum not met", "current", current, "minimum", minimum)
		return translatableerror.MinimumUAAAPIVersionNotMetError{
			Command:        command,
			MinimumVersion: minimum,
		}
	}

	return nil
}
