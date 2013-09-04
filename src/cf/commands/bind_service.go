package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
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

func (cmd *BindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement) {
	cmd.appReq = reqFactory.NewApplicationRequirement(c.String("app"))
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.String("service"))

	return []Requirement{&cmd.appReq, &cmd.serviceInstanceReq}
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
