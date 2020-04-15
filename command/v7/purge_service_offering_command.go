package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type PurgeServiceOfferingCommand struct {
	BaseCommand

	RequiredArgs    flag.Service `positional-args:"yes"`
	ServiceBroker   string       `short:"b" description:"Purge a service from a particular service broker. Required when service name is ambiguous"`
	Force           bool         `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}  `usage:"CF_NAME purge-service-offering SERVICE [-b BROKER] [-f]\n\nWARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup."`
	relatedCommands interface{}  `related_commands:"marketplace, purge-service-instance, service-brokers"`
}

func (cmd PurgeServiceOfferingCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if !cmd.Force {
		cmd.UI.DisplayText("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.\n")

		confirmed, err := cmd.confirmationPrompt()
		if err != nil {
			return err
		}
		if !confirmed {
			cmd.UI.DisplayText("Purge service offering cancelled.\n")
			return nil
		}
	}

	cmd.UI.DisplayText("Purging service offering {{.ServiceOffering}}...", map[string]interface{}{
		"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
	})

	warnings, err := cmd.Actor.PurgeServiceOfferingByNameAndBroker(cmd.RequiredArgs.ServiceOffering, cmd.ServiceBroker)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case ccerror.ServiceOfferingNotFoundError:
			cmd.UI.DisplayText("Service offering '{{.ServiceOffering}}' not found.", map[string]interface{}{
				"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd PurgeServiceOfferingCommand) confirmationPrompt() (bool, error) {
	var promptMessage string
	if cmd.ServiceBroker != "" {
		promptMessage = "Really purge service offering {{.ServiceOffering}} from broker {{.ServiceBroker}} from Cloud Foundry?"
	} else {
		promptMessage = "Really purge service offering {{.ServiceOffering}} from Cloud Foundry?"
	}

	return cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{
		"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
		"ServiceBroker":   cmd.ServiceBroker,
	})
}
