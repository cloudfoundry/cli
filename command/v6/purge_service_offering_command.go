package v6

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . PurgeServiceOfferingActor

type PurgeServiceOfferingActor interface {
	PurgeServiceOffering(service v2action.Service) (v2action.Warnings, error)
	GetServiceByNameAndBrokerName(serviceName, brokerName string) (v2action.Service, v2action.Warnings, error)
}

type PurgeServiceOfferingCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	ServiceBroker   string       `short:"b" description:"Purge a service from a particular service broker. Required when service name is ambiguous"`
	Force           bool         `short:"f" description:"Force deletion without confirmation"`
	Provider        string       `short:"p" description:"Provider"`
	usage           interface{}  `usage:"CF_NAME purge-service-offering SERVICE [-b BROKER] [-p PROVIDER] [-f]\n\nWARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup."`
	relatedCommands interface{}  `related_commands:"marketplace, purge-service-instance, service-brokers"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       PurgeServiceOfferingActor
	Config      command.Config
}

func (cmd *PurgeServiceOfferingCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd *PurgeServiceOfferingCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if cmd.Provider != "" {
		return translatableerror.FlagNoLongerSupportedError{Flag: "-p"}
	}

	service, warnings, err := cmd.Actor.GetServiceByNameAndBrokerName(cmd.RequiredArgs.ServiceOffering, cmd.ServiceBroker)
	if err != nil {
		cmd.UI.DisplayWarnings(warnings)

		switch err.(type) {
		case actionerror.ServiceNotFoundError:
			cmd.UI.DisplayText("Service offering '{{.ServiceOffering}}' not found", map[string]interface{}{
				"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
			})
			cmd.UI.DisplayOK()
			return nil
		default:
			return err
		}
	}

	cmd.UI.DisplayText("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.\n")

	if !cmd.Force {
		var promptMessage string
		if cmd.ServiceBroker != "" {
			promptMessage = "Really purge service offering {{.ServiceOffering}} from broker {{.ServiceBroker}} from Cloud Foundry?"
		} else {
			promptMessage = "Really purge service offering {{.ServiceOffering}} from Cloud Foundry?"
		}

		purgeServiceOffering, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{
			"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
			"ServiceBroker":   cmd.ServiceBroker,
		})
		if promptErr != nil {
			return promptErr
		}

		if !purgeServiceOffering {
			cmd.UI.DisplayText("Purge service offering cancelled")
			return nil
		}
	}

	cmd.UI.DisplayText("Purging service {{.ServiceOffering}}...", map[string]interface{}{
		"ServiceOffering": cmd.RequiredArgs.ServiceOffering,
	})

	purgeWarnings, err := cmd.Actor.PurgeServiceOffering(service)
	allWarnings := append(warnings, purgeWarnings...)
	cmd.UI.DisplayWarnings(allWarnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
