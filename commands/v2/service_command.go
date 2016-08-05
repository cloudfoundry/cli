package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceCommand struct {
	RequiredArgs flags.ServiceInstance `positional-args:"yes"`
	GUID         bool                  `long:"guid" description:"Retrieve and display the given service's guid.  All other output for the service is suppressed."`
}

func (_ ServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
