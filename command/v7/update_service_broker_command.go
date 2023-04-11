package v7

import (
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

	cmd.UI.DisplayTextWithFlavor(
		"Updating service broker {{.ServiceBroker}} as {{.Username}}...",
		map[string]any{
			"Username":      user.Name,
			"ServiceBroker": brokerName,
		},
	)

	warnings, err = cmd.Actor.UpdateServiceBroker(
		serviceBroker.GUID,
		resources.ServiceBroker{
			Username: username,
			Password: password,
			URL:      url,
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
