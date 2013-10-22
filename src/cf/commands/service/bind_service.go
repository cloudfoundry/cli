package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 terminal.UI
	config             *configuration.Configuration
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewBindService(ui terminal.UI, config *configuration.Configuration, serviceBindingRepo api.ServiceBindingRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceBindingRepo = serviceBindingRepo
	return
}

func (cmd *BindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "bind-service")
		return
	}
	appName := c.Args()[0]
	serviceName := c.Args()[1]

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *BindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	instance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Binding service %s to app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(instance.Name),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.serviceBindingRepo.Create(instance, app)
	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode != "90003" {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	if apiResponse.ErrorCode == "90003" {
		cmd.ui.Warn("App %s is already bound to %s.", app.Name, instance.Name)
		return
	}

	cmd.ui.Say("TIP: Use 'cf push' to ensure your env variable changes take effect")
}
