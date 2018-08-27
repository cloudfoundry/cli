package command

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	log "github.com/sirupsen/logrus"
)

func MinimumCCAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	log.WithFields(log.Fields{"current": current, "minimum": minimum}).Debug("minimum api version")
	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	isOutdated, err := checkVersionOutdated(current, minimum)
	if err != nil {
		return err
	}

	if isOutdated {
		log.WithFields(log.Fields{"current": current, "minimum": minimum}).Error("minimum not met")
		return translatableerror.MinimumCFAPIVersionNotMetError{
			Command:        command,
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}

func MinimumUAAAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	log.WithFields(log.Fields{"current": current, "minimum": minimum}).Debug("minimum api version")
	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	isOutdated, err := checkVersionOutdated(current, minimum)
	if err != nil {
		return err
	}

	if isOutdated {
		log.WithFields(log.Fields{"current": current, "minimum": minimum}).Error("minimum not met")
		return translatableerror.MinimumUAAAPIVersionNotMetError{
			Command:        command,
			MinimumVersion: minimum,
		}
	}

	return nil
}
