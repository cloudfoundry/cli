package commands

import (
	"cf/configuration"
	"cf/api"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui          term.UI
	config      *configuration.Configuration
	serviceRepo api.ServiceRepository
}

func NewDeleteService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository) (cmd DeleteService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	return
}

func (cmd DeleteService) Run(c *cli.Context) {
	instanceName := c.Args()[0]

	instance, err := cmd.serviceRepo.FindInstanceByName(cmd.config, instanceName)
	if err != nil {
		cmd.ui.Failed("", err)
		return
	}

	cmd.ui.Say("Deleting service %s...", term.Cyan(instance.Name))

	err = cmd.serviceRepo.DeleteService(cmd.config, instance)
	if err != nil {
		cmd.ui.Failed("Failed deleting service", err)
		return
	}

	cmd.ui.Ok()
}
