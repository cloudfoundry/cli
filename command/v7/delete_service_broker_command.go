package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteServiceBrokerCommand struct {
	command.BaseCommand

	RequiredArgs    flag.ServiceBroker `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME delete-service-broker SERVICE_BROKER [-f]\n\n"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	relatedCommands interface{}        `related_commands:"delete-service, purge-service-offering, service-brokers"`
}

func (cmd DeleteServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	serviceBrokerName := cmd.RequiredArgs.ServiceBroker
	if !cmd.Force {
		confirmed, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the service broker {{.ServiceBroker}}?", map[string]interface{}{
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

	cmd.UI.DisplayTextWithFlavor("Deleting service broker {{.ServiceBroker}}...",
		map[string]interface{}{
			"ServiceBroker": serviceBrokerName,
		})

	serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(serviceBrokerName)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.ServiceBrokerNotFoundError); ok {
			cmd.UI.DisplayText("Service broker '{{.ServiceBroker}}' does not exist.", map[string]interface{}{
				"ServiceBroker": serviceBrokerName,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	warnings, err = cmd.Actor.DeleteServiceBroker(serviceBroker.GUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
