package application

import (
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/uihelpers"
)

type ListApps struct {
	ui             terminal.UI
	config         coreconfig.Reader
	appSummaryRepo api.AppSummaryRepository

	pluginAppModels *[]plugin_models.GetAppsModel
	pluginCall      bool
}

func init() {
	commandregistry.Register(&ListApps{})
}

func (cmd *ListApps) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: T("List all apps in the target space"),
		Usage: []string{
			"CF_NAME apps",
		},
	}
}

func (cmd *ListApps) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

	return reqs
}

func (cmd *ListApps) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.pluginAppModels = deps.PluginModels.AppsSummary
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd *ListApps) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	apps, err := cmd.appSummaryRepo.GetSummariesInCurrentSpace()

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(apps) == 0 {
		cmd.ui.Say(T("No apps found"))
		return nil
	}

	table := cmd.ui.Table([]string{
		T("name"),
		T("requested state"),
		T("instances"),
		T("memory"),
		T("disk"),
		// Hide this column #117189491
		// T("app ports"),
		T("urls"),
	})

	for _, application := range apps {
		var urls []string
		for _, route := range application.Routes {
			urls = append(urls, route.URL())
		}

		appPorts := make([]string, len(application.AppPorts))
		for i, p := range application.AppPorts {
			appPorts[i] = strconv.Itoa(p)
		}

		table.Add(
			application.Name,
			uihelpers.ColoredAppState(application.ApplicationFields),
			uihelpers.ColoredAppInstances(application.ApplicationFields),
			formatters.ByteSize(application.Memory*formatters.MEGABYTE),
			formatters.ByteSize(application.DiskQuota*formatters.MEGABYTE),
			// Hide this column #117189491
			// strings.Join(appPorts, ", "),
			strings.Join(urls, ", "),
		)
	}

	table.Print()

	if cmd.pluginCall {
		cmd.populatePluginModel(apps)
	}
	return nil
}

func (cmd *ListApps) populatePluginModel(apps []models.Application) {
	for _, app := range apps {
		appModel := plugin_models.GetAppsModel{}
		appModel.Name = app.Name
		appModel.Guid = app.GUID
		appModel.TotalInstances = app.InstanceCount
		appModel.RunningInstances = app.RunningInstances
		appModel.Memory = app.Memory
		appModel.State = app.State
		appModel.DiskQuota = app.DiskQuota
		appModel.AppPorts = app.AppPorts

		*(cmd.pluginAppModels) = append(*(cmd.pluginAppModels), appModel)

		for _, route := range app.Routes {
			r := plugin_models.GetAppsRouteSummary{}
			r.Host = route.Host
			r.Guid = route.GUID
			r.Domain.Guid = route.Domain.GUID
			r.Domain.Name = route.Domain.Name
			r.Domain.OwningOrganizationGuid = route.Domain.OwningOrganizationGUID
			r.Domain.Shared = route.Domain.Shared

			(*(cmd.pluginAppModels))[len(*(cmd.pluginAppModels))-1].Routes = append((*(cmd.pluginAppModels))[len(*(cmd.pluginAppModels))-1].Routes, r)
		}

	}
}
