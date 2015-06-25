package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListServices struct {
	ui                 terminal.UI
	config             core_config.Reader
	serviceSummaryRepo api.ServiceSummaryRepository
	pluginModel        *[]plugin_models.GetServices_Model
	pluginCall         bool
}

func init() {
	command_registry.Register(&ListServices{})
}

func (cmd ListServices) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "services",
		ShortName:   "s",
		Description: T("List all service instances in the target space"),
		Usage:       "CF_NAME services",
	}
}

func (cmd ListServices) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("services"))
	}
	reqs = append(reqs,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	)
	return
}

func (cmd *ListServices) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceSummaryRepo = deps.RepoLocator.GetServiceSummaryRepository()
	cmd.pluginModel = deps.PluginModels.Services
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd ListServices) Execute(fc flags.FlagContext) {
	cmd.ui.Say(T("Getting services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstances, apiErr := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceInstances) == 0 {
		cmd.ui.Say(T("No services found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("service"), T("plan"), T("bound apps"), T("last operation")})

	for _, instance := range serviceInstances {
		var serviceColumn string
		var serviceStatus string

		if instance.IsUserProvided() {
			serviceColumn = T("user-provided")
		} else {
			serviceColumn = instance.ServiceOffering.Label
		}
		serviceStatus = ServiceInstanceStateToStatus(instance.LastOperation.Type, instance.LastOperation.State, instance.IsUserProvided())

		table.Add(
			instance.Name,
			serviceColumn,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
			serviceStatus,
		)
		if cmd.pluginCall {
			s := plugin_models.GetServices_Model{
				Name: instance.Name,
				Guid: instance.Guid,
				ServicePlan: plugin_models.GetServices_ServicePlan{
					Name: instance.ServicePlan.Name,
					Guid: instance.ServicePlan.Guid,
				},
				Service: plugin_models.GetServices_ServiceFields{
					Name: instance.ServiceOffering.Label,
				},
				ApplicationNames: instance.ApplicationNames,
				LastOperation: plugin_models.GetServices_LastOperation{
					Type:  instance.LastOperation.Type,
					State: instance.LastOperation.State,
				},
				IsUserProvided: instance.IsUserProvided(),
			}

			*(cmd.pluginModel) = append(*(cmd.pluginModel), s)
		}

	}

	table.Print()
}
