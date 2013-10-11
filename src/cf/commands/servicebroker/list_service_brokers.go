package servicebroker

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceBrokers struct {
	ui   terminal.UI
	repo api.ServiceBrokerRepository
}

func NewListServiceBrokers(ui terminal.UI, repo api.ServiceBrokerRepository) (cmd ListServiceBrokers) {
	cmd.ui = ui
	cmd.repo = repo
	return
}

func (cmd ListServiceBrokers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListServiceBrokers) Run(c *cli.Context) {
	cmd.ui.Say("Getting service brokers...")

	serviceBrokers, apiResponse := cmd.repo.FindAll()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceBrokers) == 0 {
		cmd.ui.Say("No service brokers found")
		return
	}

	table := [][]string{
		{"Name", "URL"},
	}

	for _, serviceBroker := range serviceBrokers {
		table = append(table, []string{
			serviceBroker.Name,
			serviceBroker.Url,
		})
	}

	cmd.ui.DisplayTable(table)
	return
}
