package service

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command *ShowService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service",
		Description: "Show service instance info",
		Usage:       "CF_NAME service SERVICE_INSTANCE",
	}
}

func (cmd *ShowService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "service")
		return
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}
	return
}

func (cmd *ShowService) Run(c *cli.Context) {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("")
	cmd.ui.Say("Service instance: %s", terminal.EntityNameColor(serviceInstance.Name))

	if serviceInstance.IsUserProvided() {
		cmd.ui.Say("Service: %s", terminal.EntityNameColor("user-provided"))
	} else {
		cmd.ui.Say("Service: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Label))
		cmd.ui.Say("Plan: %s", terminal.EntityNameColor(serviceInstance.ServicePlan.Name))
		cmd.ui.Say("Description: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.Description))
		cmd.ui.Say("Documentation url: %s", terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationUrl))
	}
}
