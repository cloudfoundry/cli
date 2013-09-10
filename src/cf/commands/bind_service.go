package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 term.UI
	serviceRepo        api.ServiceRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewBindService(ui term.UI, sR api.ServiceRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd *BindService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	appName := c.String("app")
	serviceName := c.String("service")

	if appName == "" || serviceName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "bind-service")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *BindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	instance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Binding service %s to %s...", term.Cyan(instance.Name), term.Cyan(app.Name))

	apiErr := cmd.serviceRepo.BindService(instance, app)
	if apiErr != nil && apiErr.ErrorCode != "90003" {
		cmd.ui.Failed("Failed binding service", apiErr)
		return
	}

	cmd.ui.Ok()

	if apiErr != nil && apiErr.ErrorCode == "90003" {
		cmd.ui.Say("App %s is already bound to %s.", term.Cyan(app.Name), term.Cyan(instance.Name))
	}
}
