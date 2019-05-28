package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . CreateServiceBrokerActor

type CreateServiceBrokerActor interface {
	CreateServiceBroker(v7action.ServiceBrokerCredentials) (v7action.Warnings, error)
}

type CreateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                   `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	usage           interface{}            `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"`
	relatedCommands interface{}            `related_commands:"enable-service-access, service-brokers, target"`

	SharedActor command.SharedActor
	Config      command.Config
	UI          command.UI
	Actor       CreateServiceBrokerActor
}

func (cmd *CreateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui

	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd *CreateServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Creating service broker {{.ServiceBroker}} as {{.Username}}...",
		map[string]interface{}{
			"Username":      user.Name,
			"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
		},
	)

	brokerCredentials := v7action.ServiceBrokerCredentials{
		Name:      cmd.RequiredArgs.ServiceBroker,
		URL:       cmd.RequiredArgs.URL,
		Username:  cmd.RequiredArgs.Username,
		Password:  cmd.RequiredArgs.Password,
		SpaceGUID: "",
	}

	warnings, err := cmd.Actor.CreateServiceBroker(brokerCredentials)
	cmd.UI.DisplayWarnings(warnings)

	if err == nil {
		cmd.UI.DisplayOK()
	}

	return err
}
