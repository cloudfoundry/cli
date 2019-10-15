package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateServiceBrokerActor

type CreateServiceBrokerActor interface {
	CreateServiceBroker(name, user, password, url, spaceGUID string) (v7action.Warnings, error)
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

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd *CreateServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.SpaceScoped, cmd.SpaceScoped)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	var space configv3.Space
	if cmd.SpaceScoped {
		space = cmd.Config.TargetedSpace()
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} in org {{.Org}} / space {{.Space}} as {{.Username}}...",
			map[string]interface{}{
				"Username":      user.Name,
				"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
				"Org":           cmd.Config.TargetedOrganizationName(),
				"Space":         space.Name,
			},
		)
	} else {
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} as {{.Username}}...",
			map[string]interface{}{
				"Username":      user.Name,
				"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
			},
		)
	}

	warnings, err := cmd.Actor.CreateServiceBroker(cmd.RequiredArgs.ServiceBroker, cmd.RequiredArgs.Username, cmd.RequiredArgs.Password, cmd.RequiredArgs.URL, space.GUID)
	cmd.UI.DisplayWarningsV7(warnings)

	if err == nil {
		cmd.UI.DisplayOK()
	}

	return err
}
