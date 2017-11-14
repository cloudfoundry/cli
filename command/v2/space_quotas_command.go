package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SpaceQuotasCommand struct {
	usage           interface{} `usage:"CF_NAME space-quotas"`
	relatedCommands interface{} `related_commands:"set-space-quota"`
}

func (SpaceQuotasCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SpaceQuotasCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
