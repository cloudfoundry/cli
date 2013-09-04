package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 term.UI
	config             *configuration.Configuration
	serviceRepo        api.ServiceRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewBindService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	return
}

func (cmd *BindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	appName := c.String("app")
	serviceName := c.String("service")

	if appName == "" || serviceName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "bind-service")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []Requirement{&cmd.appReq, &cmd.serviceInstanceReq}
	return
}

func (cmd *BindService) Run(c *cli.Context) {
	app := cmd.appReq.Application
	instance := cmd.serviceInstanceReq.ServiceInstance

	cmd.ui.Say("Binding service %s to %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	err := cmd.serviceRepo.BindService(cmd.config, instance, app)
	if err != nil {
		cmd.ui.Failed("Failed binding service", err)
		return
	}

	cmd.ui.Ok()
}
