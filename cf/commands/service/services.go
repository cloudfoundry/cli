package service

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/plugin/models"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type ListServices struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceSummaryRepo api.ServiceSummaryRepository
	pluginModel        *[]plugin_models.GetServices_Model
	pluginCall         bool
}

func init() {
	commandregistry.Register(&ListServices{})
}

func (cmd *ListServices) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "services",
		ShortName:   "s",
		Description: T("List all service instances in the target space"),
		Usage: []string{
			"CF_NAME services",
		},
	}
}

func (cmd *ListServices) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs, nil
}

func (cmd *ListServices) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceSummaryRepo = deps.RepoLocator.GetServiceSummaryRepository()
	cmd.pluginModel = deps.PluginModels.Services
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd *ListServices) Execute(fc flags.FlagContext) error {
	cmd.ui.Say(T("Getting services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstances, err := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceInstances) == 0 {
		cmd.ui.Say(T("No services found"))
		return nil
	}

	table := cmd.ui.Table([]string{T("name"), T("service"), T("plan"), T("bound apps"), T("last operation")})

	for _, instance := range serviceInstances {
		var serviceColumn string
		var serviceStatus string

		if instance.IsUserProvided() {
			serviceColumn = T("user-provided")
		} else {
			serviceColumn = instance.ServiceOffering.Label
		}
		serviceStatus = InstanceStateToStatus(instance.LastOperation.Type, instance.LastOperation.State, instance.IsUserProvided())

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
				Guid: instance.GUID,
				ServicePlan: plugin_models.GetServices_ServicePlan{
					Name: instance.ServicePlan.Name,
					Guid: instance.ServicePlan.GUID,
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

	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
