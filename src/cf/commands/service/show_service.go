package service

import (
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ShowService struct {
	ui                 terminal.UI
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewShowService(ui terminal.UI) (cmd *ShowService) {
	cmd = new(ShowService)
	cmd.ui = ui
	return
}

func (cmd *ShowService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "service")
		return
	}

	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}
	return
}

func (cmd *ShowService) Run(c *cli.Context) {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("")
	cmd.ui.Say("service instance: %s", terminal.EntityNameColor(serviceInstance.Name))
	cmd.ui.Say("service: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Label))
	cmd.ui.Say("plan: %s", terminal.EntityNameColor(serviceInstance.ServicePlan.Name))
	cmd.ui.Say("description: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Description))
	cmd.ui.Say("documentation url: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationUrl))
}
