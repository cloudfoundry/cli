package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RenameCommand struct {
	RequiredArgs    flag.AppRenameArgs `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME rename APP_NAME NEW_APP_NAME"`
	relatedCommands interface{}        `related_commands:"apps, delete"`
}

func (RenameCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
