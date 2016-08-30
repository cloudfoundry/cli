package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SpaceSSHAllowedCommand struct {
	RequiredArgs    flags.Space `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME space-ssh-allowed SPACE_NAME"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, ssh-enabled, ssh"`
}

func (_ SpaceSSHAllowedCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SpaceSSHAllowedCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
