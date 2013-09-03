package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui          term.UI
	config      *configuration.Configuration
	serviceRepo api.ServiceRepository
	appRepo     api.ApplicationRepository
}

func NewBindService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository, aR api.ApplicationRepository) (cmd BindService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	cmd.appRepo = aR
	return
}

func (cmd BindService) Run(c *cli.Context) {
	appName := c.String("app")
	instanceName := c.String("service")

	app, err := cmd.appRepo.FindByName(cmd.config, appName)
	if err != nil {
		cmd.ui.Failed("Application not found", err)
		return
	}

	instance, err := cmd.serviceRepo.FindInstanceByName(cmd.config, instanceName)
	if err != nil {
		cmd.ui.Failed("Service instance not found", err)
		return
	}

	cmd.ui.Say("Binding service %s to %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	err = cmd.serviceRepo.BindService(cmd.config, instance, app)
	if err != nil {
		cmd.ui.Failed("Failed binding service", err)
		return
	}

	cmd.ui.Ok()
}
