package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnbindService struct {
	ui                 terminal.UI
	config             *configuration.Configuration
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewUnbindService(ui terminal.UI, config *configuration.Configuration, serviceBindingRepo api.ServiceBindingRepository) (cmd *UnbindService) {
	cmd = new(UnbindService)
	cmd.ui = ui
	cmd.config = config
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

	cmd.ui.Say("Unbinding app %s from service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(instance.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	found, apiResponse := cmd.serviceBindingRepo.Delete(instance, app.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	if !found {
		cmd.ui.Warn("Binding between %s and %s did not exist", instance.Name, app.Name)
	}

}
