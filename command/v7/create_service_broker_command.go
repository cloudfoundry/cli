package v7

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
)

type CreateServiceBrokerCommand struct {
	BaseCommand

	PositionalArgs  flag.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                   `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	UpdateIfExists  bool                   `long:"update-if-exists" description:"If the broker already exists, update it rather than failing. Ignores --space-scoped."`
	usage           any                    `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]\n   CF_NAME create-service-broker SERVICE_BROKER USERNAME URL [--space-scoped] (omit password to specify interactively or via environment variable)\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history"`
	relatedCommands any                    `related_commands:"enable-service-access, service-brokers, target"`
	envPassword     any                    `environmentName:"CF_BROKER_PASSWORD" environmentDescription:"Password associated with user. Overridden if PASSWORD argument is provided" environmentDefault:"password"`
}

func (cmd *CreateServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.SpaceScoped, cmd.SpaceScoped)
	if err != nil {
		return err
	}

	brokerName, username, password, url, err := promptUserForBrokerPasswordIfRequired(cmd.PositionalArgs, cmd.UI)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	if cmd.UpdateIfExists {
		serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(brokerName)
		cmd.UI.DisplayWarnings(warnings)
		switch err.(type) {
		case nil:
			return updateServiceBroker(cmd.UI, cmd.Actor, user.Name, serviceBroker.GUID, brokerName, username, password, url)
		case actionerror.ServiceBrokerNotFoundError: // do nothing
		default:
			return err
		}
	}

	var space configv3.Space
	if cmd.SpaceScoped {
		space = cmd.Config.TargetedSpace()
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} in org {{.Org}} / space {{.Space}} as {{.Username}}...",
			map[string]any{
				"Username":      user.Name,
				"ServiceBroker": brokerName,
				"Org":           cmd.Config.TargetedOrganizationName(),
				"Space":         space.Name,
			},
		)
	} else {
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} as {{.Username}}...",
			map[string]any{
				"Username":      user.Name,
				"ServiceBroker": brokerName,
			},
		)
	}

	warnings, err := cmd.Actor.CreateServiceBroker(
		resources.ServiceBroker{
			Name:      brokerName,
			Username:  username,
			Password:  password,
			URL:       url,
			SpaceGUID: space.GUID,
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func promptUserForBrokerPasswordIfRequired(args flag.ServiceBrokerArgs, ui command.UI) (string, string, string, string, error) {
	if args.URL != "" {
		return args.ServiceBroker, args.Username, args.PasswordOrURL, args.URL, nil
	}

	if password, ok := os.LookupEnv("CF_BROKER_PASSWORD"); ok {
		return args.ServiceBroker, args.Username, password, args.PasswordOrURL, nil
	}

	password, err := ui.DisplayPasswordPrompt("Service Broker Password")
	if err != nil {
		return "", "", "", "", err
	}

	return args.ServiceBroker, args.Username, password, args.PasswordOrURL, nil
}
