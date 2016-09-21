package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetSpaceQuotaCommand struct {
	RequiredArgs    flags.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME set-space-quota SPACE_NAME SPACE_QUOTA_NAME"`
	relatedCommands interface{}             `related_commands:"space, space-quotas, spaces"`
}

func (_ SetSpaceQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
