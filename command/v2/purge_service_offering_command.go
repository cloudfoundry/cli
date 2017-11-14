package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type PurgeServiceOfferingCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	Force           bool         `short:"f" description:"Force deletion without confirmation"`
	Provider        string       `short:"p" description:"Provider"`
	usage           interface{}  `usage:"CF_NAME purge-service-offering SERVICE [-p PROVIDER] [-f]\n\nWARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup."`
	relatedCommands interface{}  `related_commands:"marketplace, purge-service-instance, service-brokers"`
}

func (PurgeServiceOfferingCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (PurgeServiceOfferingCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
