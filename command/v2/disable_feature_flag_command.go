package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DisableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-feature-flag FEATURE_NAME"`
	relatedCommands interface{}  `related_commands:"enable-feature-flag, feature-flags"`
}

func (DisableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DisableFeatureFlagCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
