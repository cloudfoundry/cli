package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsetSpaceQuotaCommand struct {
	RequiredArgs    flags.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME unset-space-quota SPACE SPACE_QUOTA"`
	relatedCommands interface{}             `related_commands:"space"`
}

func (_ UnsetSpaceQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UnsetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
