package application

import (
	"strings"
	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/simonleung8/flags"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/cloudfoundry/cli/flags/flag"
)

type ListApps struct {
	ui             terminal.UI
	config         core_config.Reader
	appSummaryRepo api.AppSummaryRepository

	pluginAppModels *[]plugin_models.GetAppsModel
	pluginCall      bool
	spaceRepo       spaces.SpaceRepository
}

func init() {
	command_registry.Register(&ListApps{})
}

func (cmd *ListApps) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["s"] = &cliFlags.StringFlag{Name: "s", Usage: T("Space name (list all apps in a specified space of current organization)") +
		T("\n\nTIP:\n") +
		T("By default it will list all apps of current targeted space if '-s' flag is not provided.")}

	return command_registry.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: T("List all apps in a space of targeted organization"),
		Usage:       "CF_NAME apps",
		Flags:       fs,
	}
}

func (cmd *ListApps) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("apps"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *ListApps) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.pluginAppModels = deps.PluginModels.AppsSummary
	cmd.pluginCall = pluginCall
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *ListApps) Execute(c flags.FlagContext) {

	spaceName := c.String("s")
	var apps []models.Application
	var apiErr error
	var space models.Space
	if spaceName != "" {
		space, apiErr = cmd.spaceRepo.FindByName(spaceName)
		if apiErr != nil {
			cmd.ui.Failed(T("Error finding space {{.SpaceName}}\n{{.Err}}",
				map[string]interface{}{"SpaceName": (spaceName), "Err": apiErr.Error()}))
			return
		}
		cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(space.Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))

		apps, apiErr = cmd.appSummaryRepo.GetSpaceSummaries(space.Guid)
	} else {
		cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))

		apps, apiErr = cmd.appSummaryRepo.GetSummariesInCurrentSpace()
	}

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(apps) == 0 {
		cmd.ui.Say(T("No apps found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("requested state"), T("instances"), T("memory"), T("disk"), T("urls")})

	for _, application := range apps {
		var urls []string
		for _, route := range application.Routes {
			urls = append(urls, route.URL())
		}

		table.Add(
			application.Name,
			ui_helpers.ColoredAppState(application.ApplicationFields),
			ui_helpers.ColoredAppInstances(application.ApplicationFields),
			formatters.ByteSize(application.Memory*formatters.MEGABYTE),
			formatters.ByteSize(application.DiskQuota*formatters.MEGABYTE),
			strings.Join(urls, ", "),
		)
	}

	table.Print()

	if cmd.pluginCall {
		cmd.populatePluginModel(apps)
	}
}

func (cmd *ListApps) populatePluginModel(apps []models.Application) {
	for _, app := range apps {
		appModel := plugin_models.GetAppsModel{}
		appModel.Name = app.Name
		appModel.Guid = app.Guid
		appModel.TotalInstances = app.InstanceCount
		appModel.RunningInstances = app.RunningInstances
		appModel.Memory = app.Memory
		appModel.State = app.State
		appModel.DiskQuota = app.DiskQuota

		*(cmd.pluginAppModels) = append(*(cmd.pluginAppModels), appModel)

		for _, route := range app.Routes {
			r := plugin_models.GetAppsRouteSummary{}
			r.Host = route.Host
			r.Guid = route.Guid
			r.Domain.Guid = route.Domain.Guid
			r.Domain.Name = route.Domain.Name
			r.Domain.OwningOrganizationGuid = route.Domain.OwningOrganizationGuid
			r.Domain.Shared = route.Domain.Shared

			(*(cmd.pluginAppModels))[len(*(cmd.pluginAppModels))-1].Routes = append((*(cmd.pluginAppModels))[len(*(cmd.pluginAppModels))-1].Routes, r)
		}

	}
}
