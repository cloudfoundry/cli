package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SSHEnabledCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME ssh-enabled APP_NAME"`
	relatedCommands interface{}  `related_commands:"enable-ssh, space-ssh-allowed, ssh"`
}

func (_ SSHEnabledCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SSHEnabledCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
