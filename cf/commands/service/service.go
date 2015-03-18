package service

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	. "github.com/cloudfoundry/cli/cf/i18n"
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
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service's guid.  All other output for the service is suppressed.")},
		},
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

	if c.Bool("guid") {
		cmd.ui.Say(serviceInstance.Guid)
	} else {
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
			cmd.ui.Say(T("Dashboard: {{.URL}}",
				map[string]interface{}{
					"URL": terminal.EntityNameColor(serviceInstance.DashboardUrl),
				}))
			cmd.ui.Say("")
			cmd.ui.Say(T("Last Operation"))
			cmd.ui.Say(T("Status: {{.State}}",
				map[string]interface{}{
					"State": terminal.EntityNameColor(ServiceInstanceStateToStatus(serviceInstance.LastOperation.Type, serviceInstance.LastOperation.State, serviceInstance.IsUserProvided())),
				}))
			cmd.ui.Say(T("Message: {{.Message}}",
				map[string]interface{}{
					"Message": terminal.EntityNameColor(serviceInstance.LastOperation.Description),
				}))
		}
	}
}

func ServiceInstanceStateToStatus(operationType string, state string, isUserProvidedService bool) string {
	if isUserProvidedService {
		return ""
	}

	switch state {
	case "in progress":
		return T("{{.OperationType}} in progress", map[string]interface{}{"OperationType": operationType})
	case "failed":
		return T("{{.OperationType}} failed", map[string]interface{}{"OperationType": operationType})
	case "succeeded":
		return T("{{.OperationType}} succeeded", map[string]interface{}{"OperationType": operationType})
	default:
		return ""
	}
}
