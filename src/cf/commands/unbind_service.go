package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type UnbindService struct {
	ui                 term.UI
	config             *configuration.Configuration
	serviceRepo        api.ServiceRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewUnbindService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository) (cmd *UnbindService) {
	cmd = new(UnbindService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	return
}

func (cmd *UnbindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	cmd.appReq = reqFactory.NewApplicationRequirement(c.String("app"))
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.String("service"))

	reqs = []Requirement{&cmd.appReq, &cmd.serviceInstanceReq}
	return
}

func (cmd *UnbindService) Run(c *cli.Context) {
	app := cmd.appReq.Application
	instance := cmd.serviceInstanceReq.ServiceInstance

	cmd.ui.Say("Unbinding service %s from %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	err := cmd.serviceRepo.UnbindService(cmd.config, instance, app)
	if err != nil {
		cmd.ui.Failed("Failed unbinding service", err)
		return
	}

	cmd.ui.Ok()
}
