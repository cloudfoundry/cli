package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
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
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ShowService) Run(c *cli.Context) {
	if len(c.Args()) < 1 {
		cmd.ui.FailWithUsage(c, "service")
		return
	}
	//call the CC API and get service info
	serviceInstance, _, err := cmd.serviceRepo.FindInstanceByName(c.Args()[0])

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	//TODO: check found

	cmd.ui.Say("Getting service instance %s...", terminal.EntityNameColor(serviceInstance.Name))
	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say("service instance: %s", terminal.EntityNameColor(serviceInstance.Name))
	cmd.ui.Say("service: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Label))
	cmd.ui.Say("plan: %s", terminal.EntityNameColor(serviceInstance.ServicePlan.Name))
	cmd.ui.Say("description: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Description))
	cmd.ui.Say("documentation url: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationUrl))
}
