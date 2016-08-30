package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisableSSHCommand struct {
	RequiredArgs    flags.AppName `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME disable-ssh APP_NAME"`
	relatedCommands interface{}   `related_commands:"disallow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ DisableSSHCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DisableSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
