package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CheckRouteCommand struct {
	RequiredArgs flags.HostDomain `positional-args:"yes"`
	Path         string           `long:"path" description:"Path for the route"`
}

func (_ CheckRouteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
