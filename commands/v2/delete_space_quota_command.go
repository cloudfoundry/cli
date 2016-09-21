package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteSpaceQuotaCommand struct {
	RequiredArgs    flags.SpaceQuota `positional-args:"yes"`
	Force           bool             `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}      `usage:"CF_NAME delete-space-quota SPACE_QUOTA_NAME [-f]"`
	relatedCommands interface{}      `related_commands:"space-quotas"`
}

func (_ DeleteSpaceQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
