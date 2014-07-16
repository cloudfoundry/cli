package service

import (
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

func (cmd *ShowService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service",
		Description: T("Show service instance info"),
		Usage:       T("CF_NAME service SERVICE_INSTANCE"),
	}
}

func (cmd *ShowService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
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
	cmd.ui.Say(T("Service instance: {{.ServiceName}}", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceInstance.Name)}))

	if serviceInstance.IsUserProvided() {
		cmd.ui.Say(T("Service: {{.ServiceDescription}}",
			map[string]interface{}{
				"ServiceDescription": terminal.EntityNameColor(T("user-provided")),
			}))
	} else {
		cmd.ui.Say(T("Service: {{.ServiceDescription}}",
			map[string]interface{}{
				"ServiceDescription": terminal.EntityNameColor(serviceInstance.ServiceOffering.Label),
			}))
		cmd.ui.Say(T("Plan: {{.ServicePlanName}}",
			map[string]interface{}{
				"ServicePlanName": terminal.EntityNameColor(serviceInstance.ServicePlan.Name),
			}))
		cmd.ui.Say(T("Description: {{.ServiceDescription}}", map[string]interface{}{"ServiceDescription": terminal.EntityNameColor(serviceInstance.ServiceOffering.Description)}))
		cmd.ui.Say(T("Documentation url: {{.URL}}",
			map[string]interface{}{
				"URL": terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationUrl),
			}))
	}
}
