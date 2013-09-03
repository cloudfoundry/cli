package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type UnbindService struct {
	ui          term.UI
	config      *configuration.Configuration
	serviceRepo api.ServiceRepository
	appRepo     api.ApplicationRepository
}

func NewUnbindService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository, aR api.ApplicationRepository) (cmd UnbindService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	cmd.appRepo = aR
	return
}

func (cmd UnbindService) Run(c *cli.Context) {
	appName := c.String("app")
	instanceName := c.String("service")

	app, err := cmd.appRepo.FindByName(cmd.config, appName)
	if err != nil {
		cmd.ui.Failed("", err)
		return
	}

	instance, err := cmd.serviceRepo.FindInstanceByName(cmd.config, instanceName)
	if err != nil {
		cmd.ui.Failed("", err)
		return
	}

	cmd.ui.Say("Unbinding service %s from %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	err = cmd.serviceRepo.UnbindService(cmd.config, instance, app)
	if err != nil {
		cmd.ui.Failed("Failed unbinding service", err)
		return
	}

	cmd.ui.Ok()
}
