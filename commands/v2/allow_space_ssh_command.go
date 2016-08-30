package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type AllowSpaceSSHCommand struct {
	RequiredArgs    flags.Space `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME allow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"enable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ AllowSpaceSSHCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ AllowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
