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

type ApplicationDisplayer interface {
	ShowApp(app models.Application, orgName string, spaceName string)
}

type ShowApp struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
	pluginAppModel   *plugin_models.GetAppModel
	pluginCall       bool
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
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()

	cmd.pluginAppModel = deps.PluginModels.Application
	cmd.pluginCall = pluginCall

	return cmd
}

func (cmd *ShowApp) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()

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

	if apiErr != nil && !appIsStopped {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if cmd.pluginCall {
		cmd.populatePluginModel(application, app.Stack, instances)
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
	}

	table.Print()
}

func (cmd *ShowApp) populatePluginModel(
	getSummaryApp models.Application,
	stack *models.Stack,
	instances []models.AppInstanceFields,
) {
	cmd.pluginAppModel.BuildpackUrl = getSummaryApp.BuildpackUrl
	cmd.pluginAppModel.Command = getSummaryApp.Command
	cmd.pluginAppModel.DetectedStartCommand = getSummaryApp.DetectedStartCommand
	cmd.pluginAppModel.Diego = getSummaryApp.Diego
	cmd.pluginAppModel.DiskQuota = getSummaryApp.DiskQuota
	cmd.pluginAppModel.EnvironmentVars = getSummaryApp.EnvironmentVars
	cmd.pluginAppModel.Guid = getSummaryApp.Guid
	cmd.pluginAppModel.HealthCheckTimeout = getSummaryApp.HealthCheckTimeout
	cmd.pluginAppModel.InstanceCount = getSummaryApp.InstanceCount
	cmd.pluginAppModel.Memory = getSummaryApp.Memory
	cmd.pluginAppModel.Name = getSummaryApp.Name
	cmd.pluginAppModel.PackageState = getSummaryApp.PackageState
	cmd.pluginAppModel.PackageUpdatedAt = getSummaryApp.PackageUpdatedAt
	cmd.pluginAppModel.RunningInstances = getSummaryApp.RunningInstances
	cmd.pluginAppModel.SpaceGuid = getSummaryApp.SpaceGuid
	cmd.pluginAppModel.Stack = &plugin_models.GetApp_Stack{
		Name: stack.Name,
		Guid: stack.Guid,
	}
	cmd.pluginAppModel.StagingFailedReason = getSummaryApp.StagingFailedReason
	cmd.pluginAppModel.State = getSummaryApp.State

	for _, instance := range instances {
		instanceFields := plugin_models.GetApp_AppInstanceFields{
			State:     string(instance.State),
			Details:   instance.Details,
			Since:     instance.Since,
			CpuUsage:  instance.CpuUsage,
			DiskQuota: instance.DiskQuota,
			DiskUsage: instance.DiskUsage,
			MemQuota:  instance.MemQuota,
			MemUsage:  instance.MemUsage,
		}
		cmd.pluginAppModel.Instances = append(cmd.pluginAppModel.Instances, instanceFields)
	}

	for i := range getSummaryApp.Routes {
		routeSummary := plugin_models.GetApp_RouteSummary{
			Host: getSummaryApp.Routes[i].Host,
			Guid: getSummaryApp.Routes[i].Guid,
			Domain: plugin_models.GetApp_DomainFields{
				Name: getSummaryApp.Routes[i].Domain.Name,
				Guid: getSummaryApp.Routes[i].Domain.Guid,
			},
		}
		cmd.pluginAppModel.Routes = append(cmd.pluginAppModel.Routes, routeSummary)
	}

	for i := range getSummaryApp.Services {
		serviceSummary := plugin_models.GetApp_ServiceSummary{
			Name: getSummaryApp.Services[i].Name,
			Guid: getSummaryApp.Services[i].Guid,
		}
		cmd.pluginAppModel.Services = append(cmd.pluginAppModel.Services, serviceSummary)
	}
}
