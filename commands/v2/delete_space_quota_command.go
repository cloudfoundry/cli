package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteSpaceQuotaCommand struct {
	RequiredArgs    flags.SpaceQuota `positional-args:"yes"`
	Force           bool             `short:"f" description:"Force delete (do not prompt for confirmation)"`
	usage           interface{}      `usage:"CF_NAME delete-space-quota SPACE-QUOTA-NAME [-f]"`
	relatedCommands interface{}      `related_commands:"space-quotas"`
}

func (_ DeleteSpaceQuotaCommand) Setup(config commands.Config) error {
	return nil
}

func (_ DeleteSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
