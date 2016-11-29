package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DisallowSpaceSSHCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"disable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (_ DisallowSpaceSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DisallowSpaceSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
