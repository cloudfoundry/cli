package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetSpaceQuotaCommand struct {
	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME set-space-quota SPACE_NAME SPACE_QUOTA_NAME"`
	relatedCommands interface{}            `related_commands:"space, space-quotas, spaces"`
}

func (_ SetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
