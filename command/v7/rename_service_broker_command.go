package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

type RenameServiceBrokerCommand struct {
	RequiredArgs    flag.RenameServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"`
	relatedCommands interface{}                  `related_commands:"service-brokers, update-service-broker"`

	UI          command.UI
	Config      command.Config
	Actor       UpdateServiceBrokerActor
	SharedActor command.SharedActor
}

func (cmd *RenameServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *RenameServiceBrokerCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	serviceBroker, warnings, err := cmd.Actor.GetServiceBrokerByName(cmd.RequiredArgs.OldServiceBrokerName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Renaming service broker {{.OldName}} to {{.NewName}} as {{.Username}}...",
		map[string]interface{}{
			"Username": user.Name,
			"OldName":  cmd.RequiredArgs.OldServiceBrokerName,
			"NewName":  cmd.RequiredArgs.NewServiceBrokerName,
		},
	)

	warnings, err = cmd.Actor.UpdateServiceBroker(
		serviceBroker.GUID,
		v7action.ServiceBrokerModel{
			Name: cmd.RequiredArgs.NewServiceBrokerName,
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
