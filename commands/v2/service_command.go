package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceCommand struct {
	RequiredArgs    flags.ServiceInstance `positional-args:"yes"`
	GUID            bool                  `long:"guid" description:"Retrieve and display the given service's guid.  All other output for the service is suppressed."`
	usage           interface{}           `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}           `related_commands:"bind-service, rename-service, update-service"`
}

func (_ ServiceCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ ServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
