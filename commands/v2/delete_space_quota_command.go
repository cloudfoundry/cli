package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteSpaceQuotaCommand struct {
	RequiredArgs flags.SpaceQuota `positional-args:"yes"`
	Force        bool             `short:"f" description:"Force delete (do not prompt for confirmation)"`
}

func (_ DeleteSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
