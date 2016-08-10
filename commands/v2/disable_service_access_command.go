package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisableServiceAccessCommand struct {
	RequiredArgs flags.Service `positional-args:"yes"`
	Organization string        `short:"o" description:"Disable access for a specified organization"`
	ServicePlan  string        `short:"p" description:"Disable access to a specified service plan"`
	usage        interface{}   `usage:"CF_NAME disable-service-access SERVICE [-p PLAN] [-o ORG]"`
}

func (_ DisableServiceAccessCommand) Setup() error {
	return nil
}

func (_ DisableServiceAccessCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
