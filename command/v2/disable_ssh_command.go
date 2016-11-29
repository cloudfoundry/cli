package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DisableSSHCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"disallow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ DisableSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DisableSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
