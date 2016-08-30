package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteUserCommand struct {
	RequiredArgs    flags.Username `positional-args:"yes"`
	Force           bool           `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}    `usage:"CF_NAME delete-user USERNAME [-f]"`
	relatedCommands interface{}    `related_commands:"org-users"`
}

func (_ DeleteUserCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteUserCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
