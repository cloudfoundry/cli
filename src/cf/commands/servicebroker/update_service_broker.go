package servicebroker

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateServiceBroker struct {
	ui   terminal.UI
	repo api.ServiceBrokerRepository
}

func NewUpdateServiceBroker(ui terminal.UI, repo api.ServiceBrokerRepository) (cmd UpdateServiceBroker) {
	cmd.ui = ui
	cmd.repo = repo
	return
}

func (cmd UpdateServiceBroker) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 5 {
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

	cmd.ui.Say("Updating service broker %s...", terminal.EntityNameColor(serviceBroker.Name))

	serviceBroker.Name = c.Args()[1]
	serviceBroker.Username = c.Args()[2]
	serviceBroker.Password = c.Args()[3]
	serviceBroker.Url = c.Args()[4]

	apiResponse = cmd.repo.Update(serviceBroker)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}

	cmd.ui.Ok()
}
