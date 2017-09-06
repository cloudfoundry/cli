package command

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func WarnAPIVersionCheck(config Config, ui UI) error {
	// TODO: make private and refactor commands that use
	err := MinimumAPIVersionCheck(config.BinaryVersion(), config.MinCLIVersion())

	if _, ok := err.(translatableerror.MinimumAPIVersionNotMetError); ok {
		ui.DisplayWarning("Cloud Foundry API version {{.APIVersion}} requires CLI version {{.MinCLIVersion}}. You are currently on version {{.BinaryVersion}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"APIVersion":    config.APIVersion(),
				"MinCLIVersion": config.MinCLIVersion(),
				"BinaryVersion": config.BinaryVersion(),
			})
		ui.DisplayNewline()
		return nil
	}

	// Only error if there was an issue in parsing versions
	return err
}
