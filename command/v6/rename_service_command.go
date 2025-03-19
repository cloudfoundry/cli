package v6

import (
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/flag"
	"code.cloudfoundry.org/cli/v7/command/translatableerror"
)

type RenameServiceCommand struct {
	RequiredArgs    flag.RenameServiceArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"`
	relatedCommands interface{}            `related_commands:"services, update-service"`
}

func (RenameServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameServiceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
