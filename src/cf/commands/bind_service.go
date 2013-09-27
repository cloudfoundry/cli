package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 terminal.UI
	serviceRepo        api.ServiceRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewBindService(ui terminal.UI, sR api.ServiceRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.serviceRepo = sR
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

	cmd.ui.Say("Binding service %s to %s...", terminal.EntityNameColor(instance.Name), terminal.EntityNameColor(app.Name))

	apiErr := cmd.serviceRepo.BindService(instance, app)
	if apiErr != nil && apiErr.ErrorCode != "90003" {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if apiErr != nil && apiErr.ErrorCode == "90003" {
		cmd.ui.Warn("App %s is already bound to %s.", terminal.EntityNameColor(app.Name), terminal.EntityNameColor(instance.Name))
	}
}
