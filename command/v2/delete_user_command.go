package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteUserCommand struct {
	RequiredArgs    flag.Username `positional-args:"yes"`
	Force           bool          `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}   `usage:"CF_NAME delete-user USERNAME [-f]"`
	relatedCommands interface{}   `related_commands:"org-users"`
}

func (DeleteUserCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteUserCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
