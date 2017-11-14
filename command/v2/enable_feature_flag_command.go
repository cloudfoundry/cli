package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type EnableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-feature-flag FEATURE_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, feature-flags"`
}

func (EnableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (EnableFeatureFlagCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
