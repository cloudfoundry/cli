package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsetSpaceQuotaCommand struct {
	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME unset-space-quota SPACE SPACE_QUOTA"`
	relatedCommands interface{}            `related_commands:"space"`
}

func (_ UnsetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnsetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
