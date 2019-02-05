package v6

import (
	"fmt"

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
	GetServiceByNameAndProvider(serviceName, provider string) (v2action.Service, v2action.Warnings, error)
}

type PurgeServiceOfferingCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	Force           bool         `short:"f" description:"Force deletion without confirmation"`
	Provider        string       `short:"p" description:"Provider"`
	usage           interface{}  `usage:"CF_NAME purge-service-offering SERVICE [-p PROVIDER] [-f]\n\nWARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup."`
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

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd *PurgeServiceOfferingCommand) Execute(args []string) error {
	var (
		warnings v2action.Warnings
		service  v2action.Service
		err      error
	)

	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.Provider != "" {
		if err = cmd.checkSupportedVersion(); err != nil {
			return err
		}
		service, warnings, err = cmd.Actor.GetServiceByNameAndProvider(cmd.RequiredArgs.Service, cmd.Provider)
	} else {
		service, warnings, err = cmd.Actor.GetServiceByNameAndBrokerName(cmd.RequiredArgs.Service, "")
	}

	if err != nil {
		cmd.UI.DisplayWarnings(warnings)

		switch err.(type) {
		case actionerror.ServiceNotFoundError:
			cmd.UI.DisplayText("Service offering '{{.ServiceOfferingName}}' not found.\nTIP: If you are trying to purge a v1 service offering, you must set the -p flag.", map[string]interface{}{
				"ServiceOfferingName": cmd.RequiredArgs.Service,
			})
			cmd.UI.DisplayOK()
			return nil
		default:
			return err
		}
	}

	cmd.UI.DisplayText("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.\n")

	if !cmd.Force {
		promptMessage := "Really purge service offering {{.ServiceName}} from Cloud Foundry?"
		purgeServiceOffering, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"ServiceName": cmd.RequiredArgs.Service})

		if promptErr != nil {
			return promptErr
		}

		if !purgeServiceOffering {
			cmd.UI.DisplayText("Purge service offering cancelled")
			return nil
		}
	}

	cmd.UI.DisplayText("Purging service {{.ServiceOffering}}...", map[string]interface{}{
		"ServiceOffering": cmd.RequiredArgs.Service,
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

func (cmd *PurgeServiceOfferingCommand) checkSupportedVersion() error {
	err := command.FailIfAPIVersionAboveMaxServiceProviderVersion(cmd.Config.APIVersion())
	if err != nil {
		_, ok := err.(command.APIVersionTooHighError)
		if ok {
			return fmt.Errorf("Option '-p' only works up to CF API version 2.46.0. Your target is %s.", cmd.Config.APIVersion())
		}

		return err
	}
	return nil
}
