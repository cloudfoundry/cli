package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SpaceSSHAllowedCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME space-ssh-allowed SPACE_NAME"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, ssh-enabled, ssh"`
}

func (_ SpaceSSHAllowedCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SpaceSSHAllowedCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
