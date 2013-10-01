package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ShowService struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func NewShowService(ui terminal.UI, sR api.ServiceRepository) (cmd ShowService) {
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd ShowService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "service")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ShowService) Run(c *cli.Context) {
	serviceName := c.Args()[0]

	cmd.ui.Say("Getting service instance %s...", terminal.EntityNameColor(serviceName))

	serviceInstance, apiStatus := cmd.serviceRepo.FindInstanceByName(serviceName)

	if !serviceInstance.IsFound() {
		cmd.ui.Failed("Service instance %s does not exist.", serviceName)
		return
	}

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say("service instance: %s", terminal.EntityNameColor(serviceInstance.Name))
	cmd.ui.Say("service: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Label))
	cmd.ui.Say("plan: %s", terminal.EntityNameColor(serviceInstance.ServicePlan.Name))
	cmd.ui.Say("description: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Description))
	cmd.ui.Say("documentation url: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationUrl))
}
