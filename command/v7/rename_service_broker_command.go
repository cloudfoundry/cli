package v7

import (
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type RenameServiceBrokerCommand struct {
	BaseCommand

	RequiredArgs    flag.RenameServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"`
	relatedCommands interface{}                  `related_commands:"service-brokers, update-service-broker"`
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

	user, err := cmd.Actor.GetCurrentUser()
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
		resources.ServiceBroker{
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
