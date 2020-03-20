package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . ServiceBrokersActor

type ServiceBrokersActor interface {
	GetServiceBrokers() ([]v7action.ServiceBroker, v7action.Warnings, error)
}

type ServiceBrokersCommand struct {
	usage           interface{} `usage:"CF_NAME service-brokers"`
	relatedCommands interface{} `related_commands:"delete-service-broker, disable-service-access, enable-service-access"`
	SharedActor     command.SharedActor
	Config          command.Config
	UI              command.UI
	Actor           ServiceBrokersActor
}

func (cmd *ServiceBrokersCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *ServiceBrokersCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting service brokers as {{.Username}}...", map[string]interface{}{"Username": currentUser.Name})

	serviceBrokers, warnings, err := cmd.Actor.GetServiceBrokers()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayServiceBrokers(serviceBrokers)

	return nil
}

func (cmd *ServiceBrokersCommand) displayServiceBrokers(serviceBrokers []v7action.ServiceBroker) {
	if len(serviceBrokers) == 0 {
		cmd.UI.DisplayText("No service brokers found")
	} else {
		cmd.displayServiceBrokersTable(serviceBrokers)
	}
}

func (cmd *ServiceBrokersCommand) displayServiceBrokersTable(serviceBrokers []v7action.ServiceBroker) {
	var table = [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("url"),
			cmd.UI.TranslateText("status"),
		},
	}

	for _, serviceBroker := range serviceBrokers {
		table = append(table, []string{serviceBroker.Name, serviceBroker.URL, serviceBroker.Status})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
