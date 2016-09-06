package service

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin/models"
)

type ShowService struct {
	ui                 terminal.UI
	serviceInstanceReq requirements.ServiceInstanceRequirement
	pluginModel        *plugin_models.GetService_Model
	pluginCall         bool
	appRepo            applications.Repository
}

func init() {
	commandregistry.Register(&ShowService{})
}

func (cmd *ShowService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service's guid.  All other output for the service is suppressed.")}
	T("user-provided")

	return commandregistry.CommandMetadata{
		Name:        "service",
		Description: T("Show service instance info"),
		Usage: []string{
			T("CF_NAME service SERVICE_INSTANCE"),
		},
		Flags: fs,
	}
}

func (cmd *ShowService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return reqs, nil
}

func (cmd *ShowService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI

	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Service
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()

	return cmd
}

func (cmd *ShowService) Execute(c flags.FlagContext) error {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	boundApps := []string{}
	for _, serviceBinding := range serviceInstance.ServiceBindings {
		app, err := cmd.appRepo.GetApp(serviceBinding.AppGUID)
		if err != nil {
			cmd.ui.Warn(T("Unable to retrieve information for bound application GUID " + serviceBinding.AppGUID))
		}
		boundApps = append(boundApps, app.ApplicationFields.Name)
	}

	if cmd.pluginCall {
		cmd.populatePluginModel(serviceInstance)
		return nil
	}

	if c.Bool("guid") {
		cmd.ui.Say(serviceInstance.GUID)
	} else {
		cmd.ui.Say("")
		cmd.ui.Say(T("Service instance: {{.ServiceName}}", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceInstance.Name)}))

		if serviceInstance.IsUserProvided() {
			cmd.ui.Say(T("Service: {{.ServiceDescription}}",
				map[string]interface{}{
					"ServiceDescription": terminal.EntityNameColor(T("user-provided")),
				}))
			cmd.ui.Say(T("Bound apps: {{.BoundApplications}}",
				map[string]interface{}{
					"BoundApplications": terminal.EntityNameColor(strings.Join(boundApps, ",")),
				}))
		} else {
			cmd.ui.Say(T("Service: {{.ServiceDescription}}",
				map[string]interface{}{
					"ServiceDescription": terminal.EntityNameColor(serviceInstance.ServiceOffering.Label),
				}))
			cmd.ui.Say(T("Bound apps: {{.BoundApplications}}",
				map[string]interface{}{
					"BoundApplications": terminal.EntityNameColor(strings.Join(boundApps, ",")),
				}))
			cmd.ui.Say(T("Tags: {{.Tags}}",
				map[string]interface{}{
					"Tags": terminal.EntityNameColor(strings.Join(serviceInstance.Tags, ", ")),
				}))
			cmd.ui.Say(T("Plan: {{.ServicePlanName}}",
				map[string]interface{}{
					"ServicePlanName": terminal.EntityNameColor(serviceInstance.ServicePlan.Name),
				}))
			cmd.ui.Say(T("Description: {{.ServiceDescription}}", map[string]interface{}{"ServiceDescription": terminal.EntityNameColor(serviceInstance.ServiceOffering.Description)}))
			cmd.ui.Say(T("Documentation url: {{.URL}}",
				map[string]interface{}{
					"URL": terminal.EntityNameColor(serviceInstance.ServiceOffering.DocumentationURL),
				}))
			cmd.ui.Say(T("Dashboard: {{.URL}}",
				map[string]interface{}{
					"URL": terminal.EntityNameColor(serviceInstance.DashboardURL),
				}))
			cmd.ui.Say("")
			cmd.ui.Say(T("Last Operation"))
			cmd.ui.Say(T("Status: {{.State}}",
				map[string]interface{}{
					"State": terminal.EntityNameColor(InstanceStateToStatus(serviceInstance.LastOperation.Type, serviceInstance.LastOperation.State, serviceInstance.IsUserProvided())),
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
	return nil
}

func InstanceStateToStatus(operationType string, state string, isUserProvidedService bool) string {
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

func (cmd *ShowService) populatePluginModel(serviceInstance models.ServiceInstance) {
	cmd.pluginModel.Name = serviceInstance.Name
	cmd.pluginModel.Guid = serviceInstance.GUID
	cmd.pluginModel.DashboardUrl = serviceInstance.DashboardURL
	cmd.pluginModel.IsUserProvided = serviceInstance.IsUserProvided()
	cmd.pluginModel.LastOperation.Type = serviceInstance.LastOperation.Type
	cmd.pluginModel.LastOperation.State = serviceInstance.LastOperation.State
	cmd.pluginModel.LastOperation.Description = serviceInstance.LastOperation.Description
	cmd.pluginModel.LastOperation.CreatedAt = serviceInstance.LastOperation.CreatedAt
	cmd.pluginModel.LastOperation.UpdatedAt = serviceInstance.LastOperation.UpdatedAt
	cmd.pluginModel.ServicePlan.Name = serviceInstance.ServicePlan.Name
	cmd.pluginModel.ServicePlan.Guid = serviceInstance.ServicePlan.GUID
	cmd.pluginModel.ServiceOffering.DocumentationUrl = serviceInstance.ServiceOffering.DocumentationURL
	cmd.pluginModel.ServiceOffering.Name = serviceInstance.ServiceOffering.Label
}
