package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type FeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME feature-flag FEATURE_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, enable-feature-flag, feature-flags"`
}

func (FeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (FeatureFlagCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
