package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetSpaceQuotaCommand struct {
	RequiredArgs flags.SetSpaceQuotaArgs `positional-args:"yes"`
}

func (_ SetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
