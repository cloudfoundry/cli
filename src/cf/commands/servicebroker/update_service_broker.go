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
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewUpdateServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd UpdateServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd UpdateServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "update-service-broker")
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())

	return
}

func (cmd UpdateServiceBroker) Run(c *cli.Context) {
	serviceBroker, apiErr := cmd.repo.FindByName(c.Args()[0])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say("Updating service broker %s as %s...",
		terminal.EntityNameColor(serviceBroker.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceBroker.Username = c.Args()[1]
	serviceBroker.Password = c.Args()[2]
	serviceBroker.Url = c.Args()[3]

	apiErr = cmd.repo.Update(serviceBroker)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
