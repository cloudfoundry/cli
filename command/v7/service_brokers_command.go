package v7

import (
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type ServiceBrokersCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME service-brokers"`
	relatedCommands interface{} `related_commands:"delete-service-broker, disable-service-access, enable-service-access"`
}

func (cmd *ServiceBrokersCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Actor.GetCurrentUser()
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

func (cmd *ServiceBrokersCommand) displayServiceBrokers(serviceBrokers []resources.ServiceBroker) {
	if len(serviceBrokers) == 0 {
		cmd.UI.DisplayText("No service brokers found")
	} else {
		cmd.displayServiceBrokersTable(serviceBrokers)
	}
}

func (cmd *ServiceBrokersCommand) displayServiceBrokersTable(serviceBrokers []resources.ServiceBroker) {
	var table = [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("url"),
		},
	}

	for _, serviceBroker := range serviceBrokers {
		table = append(table, []string{serviceBroker.Name, serviceBroker.URL})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
