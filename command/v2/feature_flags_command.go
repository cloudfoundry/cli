package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type FeatureFlagsCommand struct {
	usage           interface{} `usage:"CF_NAME feature-flags"`
	relatedCommands interface{} `related_commands:"disable-feature-flag, enable-feature-flag"`
}

func (FeatureFlagsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (FeatureFlagsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
