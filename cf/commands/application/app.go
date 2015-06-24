package application

import (
	"fmt"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
)

type ShowApp struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	appLogsNoaaRepo  api.LogsNoaaRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
	pluginAppModel   *plugin_models.GetAppModel
	pluginCall       bool
}

type ApplicationDisplayer interface {
	ShowApp(app models.Application, orgName string, spaceName string)
}

func NewShowApp(ui terminal.UI, config core_config.Reader, appSummaryRepo api.AppSummaryRepository, appInstancesRepo app_instances.AppInstancesRepository, appLogsNoaaRepo api.LogsNoaaRepository) (cmd *ShowApp) {
	cmd = &ShowApp{}
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	cmd.appInstancesRepo = appInstancesRepo
	cmd.appLogsNoaaRepo = appLogsNoaaRepo
	return
}

func init() {
	command_registry.Register(&ShowApp{})
}

func (cmd *ShowApp) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &cliFlags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given app's guid.  All other health and status output for the app is suppressed.")}

	return command_registry.CommandMetadata{
		Name:        "app",
		Description: T("Display health and status for app"),
		Usage:       T("CF_NAME app APP_NAME"),
		Flags:       fs,
	}
}

func (cmd *ShowApp) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("app"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *ShowApp) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.appLogsNoaaRepo = deps.RepoLocator.GetLogsNoaaRepository()
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()

	cmd.pluginAppModel = deps.PluginModels.Application
	cmd.pluginCall = pluginCall

	return cmd
}

func (cmd *ShowApp) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	if cmd.pluginCall {
		cmd.pluginAppModel.Name = app.Name
		cmd.pluginAppModel.State = app.State
		cmd.pluginAppModel.Guid = app.Guid
		cmd.pluginAppModel.BuildpackUrl = app.BuildpackUrl
		cmd.pluginAppModel.Command = app.Command
		cmd.pluginAppModel.Diego = app.Diego
		cmd.pluginAppModel.DetectedStartCommand = app.DetectedStartCommand
		cmd.pluginAppModel.DiskQuota = app.DiskQuota
		cmd.pluginAppModel.EnvironmentVars = app.EnvironmentVars
		cmd.pluginAppModel.InstanceCount = app.InstanceCount
		cmd.pluginAppModel.Memory = app.Memory
		cmd.pluginAppModel.RunningInstances = app.RunningInstances
		cmd.pluginAppModel.HealthCheckTimeout = app.HealthCheckTimeout
		cmd.pluginAppModel.SpaceGuid = app.SpaceGuid
		cmd.pluginAppModel.PackageUpdatedAt = app.PackageUpdatedAt
		cmd.pluginAppModel.PackageState = app.PackageState
		cmd.pluginAppModel.StagingFailedReason = app.StagingFailedReason

		cmd.pluginAppModel.Stack = &plugin_models.GetApp_Stack{
			Name: app.Stack.Name,
			Guid: app.Stack.Guid,
		}

		for i, _ := range app.Routes {
			cmd.pluginAppModel.Routes = append(cmd.pluginAppModel.Routes, plugin_models.GetApp_RouteSummary{
				Host: app.Routes[i].Host,
				Guid: app.Routes[i].Guid,
				Domain: plugin_models.DomainFields{
					Name:                   app.Routes[i].Domain.Name,
					Guid:                   app.Routes[i].Domain.Guid,
					Shared:                 app.Routes[i].Domain.Shared,
					OwningOrganizationGuid: app.Routes[i].Domain.OwningOrganizationGuid,
				},
			})
		}

		for i, _ := range app.Services {
			cmd.pluginAppModel.Services = append(cmd.pluginAppModel.Services, plugin_models.GetApp_ServiceSummary{
				Name: app.Services[i].Name,
				Guid: app.Services[i].Guid,
			})
		}
	}

	if c.Bool("guid") {
		cmd.ui.Say(app.Guid)
	} else {
		cmd.ShowApp(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}
}

func (cmd *ShowApp) ShowApp(app models.Application, orgName, spaceName string) {
	cmd.ui.Say(T("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(orgName),
			"SpaceName": terminal.EntityNameColor(spaceName),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	application, apiErr := cmd.appSummaryRepo.GetSummary(app.Guid)

	appIsStopped := (application.State == "stopped")
	if err, ok := apiErr.(errors.HttpError); ok {
		if err.ErrorCode() == errors.APP_STOPPED || err.ErrorCode() == errors.APP_NOT_STAGED {
			appIsStopped = true
		}
	}

	if apiErr != nil && !appIsStopped {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	var instances []models.AppInstanceFields
	instances, apiErr = cmd.appInstancesRepo.GetInstances(app.Guid)
	if apiErr != nil && !appIsStopped {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	//temp solution, diego app metrics only come from noaa, not CC
	if application.Diego {
		instances, apiErr = cmd.appLogsNoaaRepo.GetContainerMetrics(app.Guid, instances)

		for i := 0; i < len(instances); i++ {
			instances[i].MemQuota = application.Memory * 1024 * 1024
			instances[i].DiskQuota = application.DiskQuota * 1024 * 1024
		}
	}

	if apiErr != nil && !appIsStopped {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", terminal.HeaderColor(T("requested state:")), ui_helpers.ColoredAppState(application.ApplicationFields))
	cmd.ui.Say("%s %s", terminal.HeaderColor(T("instances:")), ui_helpers.ColoredAppInstances(application.ApplicationFields))
	cmd.ui.Say(T("{{.Usage}} {{.FormattedMemory}} x {{.InstanceCount}} instances",
		map[string]interface{}{
			"Usage":           terminal.HeaderColor(T("usage:")),
			"FormattedMemory": formatters.ByteSize(application.Memory * formatters.MEGABYTE),
			"InstanceCount":   application.InstanceCount}))

	var urls []string
	for _, route := range application.Routes {
		urls = append(urls, route.URL())
	}

	cmd.ui.Say("%s %s", terminal.HeaderColor(T("urls:")), strings.Join(urls, ", "))
	var lastUpdated string
	if application.PackageUpdatedAt != nil {
		lastUpdated = application.PackageUpdatedAt.Format("Mon Jan 2 15:04:05 MST 2006")
	} else {
		lastUpdated = "unknown"
	}
	cmd.ui.Say("%s %s", terminal.HeaderColor(T("last uploaded:")), lastUpdated)
	if app.Stack != nil {
		cmd.ui.Say("%s %s", terminal.HeaderColor(T("stack:")), app.Stack.Name)
	} else {
		cmd.ui.Say("%s %s", terminal.HeaderColor(T("stack:")), "unknown")
	}

	if app.Buildpack != "" {
		cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("buildpack:")), app.Buildpack)
	} else if app.DetectedBuildpack != "" {
		cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("buildpack:")), app.DetectedBuildpack)
	} else {
		cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("buildpack:")), "unknown")
	}

	if appIsStopped {
		cmd.ui.Say(T("There are no running instances of this app."))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{"", T("state"), T("since"), T("cpu"), T("memory"), T("disk"), T("details")})

	for index, instance := range instances {
		table.Add(
			fmt.Sprintf("#%d", index),
			ui_helpers.ColoredInstanceState(instance),
			instance.Since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", instance.CpuUsage*100),
			fmt.Sprintf(T("{{.MemUsage}} of {{.MemQuota}}",
				map[string]interface{}{
					"MemUsage": formatters.ByteSize(instance.MemUsage),
					"MemQuota": formatters.ByteSize(instance.MemQuota)})),
			fmt.Sprintf(T("{{.DiskUsage}} of {{.DiskQuota}}",
				map[string]interface{}{
					"DiskUsage": formatters.ByteSize(instance.DiskUsage),
					"DiskQuota": formatters.ByteSize(instance.DiskQuota)})),
			fmt.Sprintf("%s", instance.Details),
		)

		if cmd.pluginCall {
			i := plugin_models.GetApp_AppInstanceFields{}
			i.State = fmt.Sprintf("%s", instance.State)
			i.Details = instance.Details
			i.Since = instance.Since
			i.CpuUsage = instance.CpuUsage
			i.DiskQuota = instance.DiskQuota
			i.DiskUsage = instance.DiskUsage
			i.MemQuota = instance.MemQuota
			i.MemUsage = instance.MemUsage
			cmd.pluginAppModel.Instances = append(cmd.pluginAppModel.Instances, i)
		}
	}

	table.Print()
}
