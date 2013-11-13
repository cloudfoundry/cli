package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceBrokers struct {
	ui     terminal.UI
	config *configuration.Configuration
	repo   api.ServiceBrokerRepository
}

func NewListServiceBrokers(ui terminal.UI, config *configuration.Configuration, repo api.ServiceBrokerRepository) (cmd ListServiceBrokers) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd ListServiceBrokers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListServiceBrokers) Run(c *cli.Context) {
	cmd.ui.Say("Getting service brokers as %s...\n", terminal.EntityNameColor(cmd.config.Username()))

	stopChan := make(chan bool)
	defer close(stopChan)

	serviceBrokersChan, statusChan := cmd.repo.ListServiceBrokers(stopChan)

	table := cmd.ui.Table([]string{"name", "url"})
	noServiceBrokers := true

	for serviceBrokers := range serviceBrokersChan {
		rows := [][]string{}
		for _, serviceBroker := range serviceBrokers {
			rows = append(rows, []string{
				serviceBroker.Name,
				serviceBroker.Url,
			})
		}
		table.Print(rows)
		noServiceBrokers = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching service brokers.\n%s", apiStatus.Message)
		return
	}

	if noServiceBrokers {
		cmd.ui.Say("No service brokers found")
	}
}
