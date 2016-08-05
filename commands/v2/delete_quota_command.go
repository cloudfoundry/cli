package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteQuotaCommand struct {
	RequiredArgs flags.Quota `positional-args:"yes"`
	Force        bool        `short:"f" description:"Force deletion without confirmation"`
}

func (_ DeleteQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
