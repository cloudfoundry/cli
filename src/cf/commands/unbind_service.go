package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnbindService struct {
	ui                 term.UI
	serviceRepo        api.ServiceRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewUnbindService(ui term.UI, sR api.ServiceRepository) (cmd *UnbindService) {
	cmd = new(UnbindService)
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd *UnbindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	appName := c.String("app")
	serviceName := c.String("service")

	if appName == "" || serviceName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "unbind-service")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *UnbindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	instance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Unbinding service %s from %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	err := cmd.serviceRepo.UnbindService(instance, app)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
