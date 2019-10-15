package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UpdateServiceBrokerActor

type UpdateServiceBrokerActor interface {
	GetServiceBrokerByName(serviceBrokerName string) (v7action.ServiceBroker, v7action.Warnings, error)
	UpdateServiceBroker(serviceBrokerGUID, name, user, password, url string) (v7action.Warnings, error)
}

type UpdateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"`
	relatedCommands interface{}            `related_commands:"rename-service-broker, service-brokers"`

	UI          command.UI
	Config      command.Config
	Actor       UpdateServiceBrokerActor
	SharedActor command.SharedActor
}

func (cmd *UpdateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	sharedActor := sharedaction.NewActor(config)
	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedActor
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
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

	warnings, err = cmd.Actor.UpdateServiceBroker(
		serviceBroker.GUID,
		"",
		cmd.RequiredArgs.Username,
		cmd.RequiredArgs.Password,
		cmd.RequiredArgs.URL,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	return nil
}
