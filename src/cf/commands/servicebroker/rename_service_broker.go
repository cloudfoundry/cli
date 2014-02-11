package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameServiceBroker struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewRenameServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd RenameServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd RenameServiceBroker) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-service-broker")
		return
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())

	return
}

func (cmd RenameServiceBroker) Run(c *cli.Context) {
	serviceBroker, apiResponse := cmd.repo.FindByName(c.Args()[0])
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Say("Renaming service broker %s to %s as %s",
		terminal.EntityNameColor(serviceBroker.Name),
		terminal.EntityNameColor(c.Args()[1]),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	newName := c.Args()[1]

	apiResponse = cmd.repo.Rename(serviceBroker.Guid, newName)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
