package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type PurgeServiceInstanceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}          `usage:"CF_NAME purge-service-instance SERVICE_INSTANCE\n\nWARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys."`
	relatedCommands interface{}          `related_commands:"delete-service, services, service-brokers"`
}

func (_ PurgeServiceInstanceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ PurgeServiceInstanceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
