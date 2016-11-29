package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableSSHCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"allow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ EnableSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ EnableSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
