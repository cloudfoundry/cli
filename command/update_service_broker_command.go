package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
)

type UpdateServiceBrokerCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"`
	relatedCommands interface{}            `related_commands:"rename-service-broker, service-brokers"`
}

func (cmd UpdateServiceBrokerCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(cmd.RequiredArgs.ServiceBroker)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Updating service broker {{.ServiceBroker}} as {{.Username}}...",
		map[string]interface{}{
			"Username":      user.Name,
			"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
		},
	)

	warnings, err = cmd.Actor.UpdateServiceBroker(
		serviceBroker.GUID,
		resources.ServiceBroker{
			Username: cmd.RequiredArgs.Username,
			Password: cmd.RequiredArgs.Password,
			URL:      cmd.RequiredArgs.URL,
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
