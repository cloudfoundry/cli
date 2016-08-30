package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisallowSpaceSSHCommand struct {
	RequiredArgs    flags.Space `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"disable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ DisallowSpaceSSHCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DisallowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
