package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisableServiceAccessCommand struct {
	RequiredArgs flags.Service `positional-args:"yes"`
	ServicePlan  string        `short:"p" description:"Enable access to a specified service plan"`
	Organization string        `short:"o" description:"Enable access to a specified organization"`
}

func (_ DisableServiceAccessCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
