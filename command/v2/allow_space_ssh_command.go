package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type AllowSpaceSSHCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME allow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"enable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ AllowSpaceSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ AllowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
