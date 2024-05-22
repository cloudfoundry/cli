package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
)

type UpdateServiceBrokerCommand struct {
	BaseCommand

	PositionalArgs  flag.ServiceBrokerArgs `positional-args:"yes"`
	usage           any                    `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL\n   CF_NAME update-service-broker SERVICE_BROKER USERNAME URL (omit password to specify interactively or via environment variable)\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history"`
	relatedCommands any                    `related_commands:"rename-service-broker, service-brokers"`
	envPassword     any                    `environmentName:"CF_BROKER_PASSWORD" environmentDescription:"Password associated with user. Overridden if PASSWORD argument is provided" environmentDefault:"password"`
}

func (cmd UpdateServiceBrokerCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	brokerName, username, password, url, err := promptUserForBrokerPasswordIfRequired(cmd.PositionalArgs, cmd.UI)
	if err != nil {
		return err
	}

	serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(brokerName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	return updateServiceBroker(cmd.UI, cmd.Actor, user.Name, serviceBroker.GUID, brokerName, username, password, url)
}

func updateServiceBroker(ui command.UI, actor Actor, user, brokerGUID, brokerName, username, password, url string) error {
	ui.DisplayTextWithFlavor(
		"Updating service broker {{.ServiceBroker}} as {{.Username}}...",
		map[string]any{
			"Username":      user,
			"ServiceBroker": brokerName,
		},
	)

	warnings, err := actor.UpdateServiceBroker(
		brokerGUID,
		resources.ServiceBroker{
			Username: username,
			Password: password,
			URL:      url,
		},
	)
	ui.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	ui.DisplayOK()

	return nil
}
