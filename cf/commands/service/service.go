package service

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type ShowService struct {
	ui                 terminal.UI
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&ShowService{})
}

func (cmd *ShowService) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &cliFlags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service's guid.  All other output for the service is suppressed.")}
	return command_registry.CommandMetadata{
		Name:        "service",
		Description: T("Show service instance info"),
		Usage:       T("CF_NAME service SERVICE_INSTANCE"),
		Flags:       fs,
	}
}

func (cmd *ShowService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("service"))
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *ShowService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	return cmd
}

func (cmd *ShowService) Execute(c flags.FlagContext) {
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
			if "" != serviceInstance.LastOperation.CreatedAt {
				cmd.ui.Say(T("Started: {{.Started}}",
					map[string]interface{}{
						"Started": terminal.EntityNameColor(serviceInstance.LastOperation.CreatedAt),
					}))
			}
			cmd.ui.Say(T("Updated: {{.Updated}}",
				map[string]interface{}{
					"Updated": terminal.EntityNameColor(serviceInstance.LastOperation.UpdatedAt),
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
