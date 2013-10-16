package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnbindService struct {
	ui                 terminal.UI
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewUnbindService(ui terminal.UI, serviceBindingRepo api.ServiceBindingRepository) (cmd *UnbindService) {
	cmd = new(UnbindService)
	cmd.ui = ui
	cmd.serviceBindingRepo = serviceBindingRepo
	return
}

func (cmd *UnbindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "unbind-service")
		return
	}

	appName := c.Args()[0]
	serviceName := c.Args()[1]

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *UnbindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	instance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Unbinding service %s from %s...", terminal.EntityNameColor(instance.Name), terminal.EntityNameColor(app.Name))

	found, apiResponse := cmd.serviceBindingRepo.Delete(instance, app)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	if !found {
		cmd.ui.Warn("Binding between %s and %s did not exist", instance.Name, app.Name)
	}

}
