package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceKeyCommand struct {
	RequiredArgs flags.ServiceInstanceKey `positional-args:"yes"`
	GUID         bool                     `long:"guid" description:"Retrieve and display the given service-key's guid.  All other output for the service is suppressed."`
}

func (_ ServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
