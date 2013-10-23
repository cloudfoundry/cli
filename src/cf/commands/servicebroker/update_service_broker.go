package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateServiceBroker struct {
	ui     terminal.UI
	config *configuration.Configuration
	repo   api.ServiceBrokerRepository
}

func NewUpdateServiceBroker(ui terminal.UI, config *configuration.Configuration, repo api.ServiceBrokerRepository) (cmd UpdateServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd UpdateServiceBroker) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "update-service-broker")
		return
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())

	return
}

func (cmd UpdateServiceBroker) Run(c *cli.Context) {
	serviceBroker, apiResponse := cmd.repo.FindByName(c.Args()[0])
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Say("Updating service broker %s as %s...",
		terminal.EntityNameColor(serviceBroker.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceBroker.Username = c.Args()[1]
	serviceBroker.Password = c.Args()[2]
	serviceBroker.Url = c.Args()[3]

	apiResponse = cmd.repo.Update(serviceBroker)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
