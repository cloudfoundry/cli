package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SpaceQuotaCommand struct {
	RequiredArgs flags.SpaceQuota `positional-args:"yes"`
	usage        interface{}      `usage:"CF_NAME space-quota SPACE_QUOTA_NAME"`
}

func (_ SpaceQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
