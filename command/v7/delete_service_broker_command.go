package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UpdateServiceBrokerActor

type DeleteServiceBrokerActor interface {
	GetServiceBrokerByName(serviceBrokerName string) (v7action.ServiceBroker, v7action.Warnings, error)
	DeleteServiceBroker(serviceBrokerGUID string) (v7action.Warnings, error)
}

type DeleteServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBroker `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME delete-service-broker SERVICE_BROKER [-f]\n\n"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	relatedCommands interface{}        `related_commands:"delete-service, purge-service-offering, service-brokers"`

	UI          command.UI
	Config      command.Config
	Actor       DeleteServiceBrokerActor
	SharedActor command.SharedActor
}

func (cmd *DeleteServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
}

func (cmd DeleteServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	serviceBrokerName := cmd.RequiredArgs.ServiceBroker
	if !cmd.Force {
		confirmed, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the service-broker {{.ServiceBroker}}?", map[string]interface{}{
			"ServiceBroker": serviceBrokerName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !confirmed {
			cmd.UI.DisplayText("'{{.ServiceBroker}}' has not been deleted.", map[string]interface{}{
				"ServiceBroker": serviceBrokerName,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting service-broker {{.ServiceBroker}}...",
		map[string]interface{}{
			"ServiceBroker": serviceBrokerName,
		})

	serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(serviceBrokerName)

	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		if _, ok := err.(actionerror.ServiceBrokerNotFoundError); ok {
			// TODO: Verify the correct error message to display for idempotent case
			//cmd.UI.DisplayText(`Unable to delete. ` + err.Error())
			cmd.UI.DisplayText("Service broker '{{.ServiceBroker}}' does not exist.", map[string]interface{}{
				"ServiceBroker": serviceBrokerName,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	warnings, err = cmd.Actor.DeleteServiceBroker(serviceBroker.GUID)

	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
