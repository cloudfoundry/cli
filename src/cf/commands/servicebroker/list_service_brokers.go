package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceBrokers struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewListServiceBrokers(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd ListServiceBrokers) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd ListServiceBrokers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd ListServiceBrokers) Run(c *cli.Context) {
	cmd.ui.Say("Getting service brokers as %s...\n", terminal.EntityNameColor(cmd.config.Username()))

	table := cmd.ui.Table([]string{"name", "url"})
	foundBrokers := false
	apiStatus := cmd.repo.ListServiceBrokers(func(serviceBroker models.ServiceBroker) bool {
		table.Print([][]string{{serviceBroker.Name, serviceBroker.Url}})
		foundBrokers = true
		return true
	})

	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching service brokers.\n%s", apiStatus.Message)
		return
	}

	if !foundBrokers {
		cmd.ui.Say("No service brokers found")
	}
}
