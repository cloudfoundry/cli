package application

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/plugin/models"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/stacks"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/formatters"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/uihelpers"
)

//go:generate counterfeiter . Displayer

type Displayer interface {
	ShowApp(app models.Application, orgName string, spaceName string) error
}

type ShowApp struct {
	ui               terminal.UI
	config           coreconfig.Reader
	appSummaryRepo   api.AppSummaryRepository
	appInstancesRepo appinstances.Repository
	stackRepo        stacks.StackRepository
	appReq           requirements.ApplicationRequirement
	pluginAppModel   *plugin_models.GetAppModel
	pluginCall       bool
}

func init() {
	commandregistry.Register(&ShowApp{})
}

func (cmd *ShowApp) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given app's guid.  All other health and status output for the app is suppressed.")}

	return commandregistry.CommandMetadata{
		Name:        "app",
		Description: T("Display health and status for app"),
		Usage: []string{
			T("CF_NAME app APP_NAME"),
		},
		Flags: fs,
	}
}

func (cmd *ShowApp) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("app"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *ShowApp) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()

	cmd.pluginAppModel = deps.PluginModels.Application
	cmd.pluginCall = pluginCall

	return cmd
}

func (cmd *ShowApp) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	if c.Bool("guid") {
		cmd.ui.Say(app.GUID)
	} else {
		err := cmd.ShowApp(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd *ShowApp) ShowApp(app models.Application, orgName, spaceName string) error {
	cmd.ui.Say(T("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(orgName),
			"SpaceName": terminal.EntityNameColor(spaceName),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	application, err := cmd.appSummaryRepo.GetSummary(app.GUID)

	appIsStopped := (application.State == "stopped")
	if assertionErr, ok := err.(errors.HTTPError); ok {
		if assertionErr.ErrorCode() == errors.InstancesError || assertionErr.ErrorCode() == errors.NotStaged {
			appIsStopped = true
		}
	}

	if err != nil && !appIsStopped {
		return err
	}

	var instances []models.AppInstanceFields
	instances, err = cmd.appInstancesRepo.GetInstances(app.GUID)
	if err != nil && !appIsStopped {
		return err
	}

	if cmd.pluginCall {
		cmd.populatePluginModel(application, app.Stack, instances)
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", terminal.HeaderColor(T("requested state:")), uihelpers.ColoredAppState(application.ApplicationFields))
	cmd.ui.Say("%s %s", terminal.HeaderColor(T("instances:")), uihelpers.ColoredAppInstances(application.ApplicationFields))

	// Commented to hide app-ports for release #117189491
	// if len(application.AppPorts) > 0 {
	// 	appPorts := make([]string, len(application.AppPorts))
	// 	for i, p := range application.AppPorts {
	// 		appPorts[i] = strconv.Itoa(p)
	// 	}

	// 	cmd.ui.Say("%s %s", terminal.HeaderColor(T("app ports:")), strings.Join(appPorts, ", "))
	// }

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

	appStack, err := cmd.stackRepo.FindByGUID(application.ApplicationFields.StackGUID)
	if appStack.Name != "" && err == nil {
		cmd.ui.Say("%s %s", terminal.HeaderColor(T("stack:")), appStack.Name)
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
		return nil
	}

	table := cmd.ui.Table([]string{"", T("state"), T("since"), T("cpu"), T("memory"), T("disk"), T("details")})

	for index, instance := range instances {
		table.Add(
			fmt.Sprintf("#%d", index),
			uihelpers.ColoredInstanceState(instance),
			instance.Since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", instance.CPUUsage*100),
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

	err = table.Print()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

func (cmd *ShowApp) populatePluginModel(
	getSummaryApp models.Application,
	stack *models.Stack,
	instances []models.AppInstanceFields,
) {
	cmd.pluginAppModel.BuildpackUrl = getSummaryApp.BuildpackURL
	cmd.pluginAppModel.Command = getSummaryApp.Command
	cmd.pluginAppModel.DetectedStartCommand = getSummaryApp.DetectedStartCommand
	cmd.pluginAppModel.Diego = getSummaryApp.Diego
	cmd.pluginAppModel.DiskQuota = getSummaryApp.DiskQuota
	cmd.pluginAppModel.EnvironmentVars = getSummaryApp.EnvironmentVars
	cmd.pluginAppModel.Guid = getSummaryApp.GUID
	cmd.pluginAppModel.HealthCheckTimeout = getSummaryApp.HealthCheckTimeout
	cmd.pluginAppModel.InstanceCount = getSummaryApp.InstanceCount
	cmd.pluginAppModel.Memory = getSummaryApp.Memory
	cmd.pluginAppModel.Name = getSummaryApp.Name
	cmd.pluginAppModel.PackageState = getSummaryApp.PackageState
	cmd.pluginAppModel.PackageUpdatedAt = getSummaryApp.PackageUpdatedAt
	cmd.pluginAppModel.RunningInstances = getSummaryApp.RunningInstances
	cmd.pluginAppModel.SpaceGuid = getSummaryApp.SpaceGUID
	cmd.pluginAppModel.AppPorts = getSummaryApp.AppPorts
	cmd.pluginAppModel.Stack = &plugin_models.GetApp_Stack{
		Name: stack.Name,
		Guid: stack.GUID,
	}
	cmd.pluginAppModel.StagingFailedReason = getSummaryApp.StagingFailedReason
	cmd.pluginAppModel.State = getSummaryApp.State

	for _, instance := range instances {
		instanceFields := plugin_models.GetApp_AppInstanceFields{
			State:     string(instance.State),
			Details:   instance.Details,
			Since:     instance.Since,
			CpuUsage:  instance.CPUUsage,
			DiskQuota: instance.DiskQuota,
			DiskUsage: instance.DiskUsage,
			MemQuota:  instance.MemQuota,
			MemUsage:  instance.MemUsage,
		}
		cmd.pluginAppModel.Instances = append(cmd.pluginAppModel.Instances, instanceFields)
	}
	if cmd.pluginAppModel.Instances == nil {
		cmd.pluginAppModel.Instances = []plugin_models.GetApp_AppInstanceFields{}
	}

	for i := range getSummaryApp.Routes {
		routeSummary := plugin_models.GetApp_RouteSummary{
			Host: getSummaryApp.Routes[i].Host,
			Guid: getSummaryApp.Routes[i].GUID,
			Domain: plugin_models.GetApp_DomainFields{
				Name: getSummaryApp.Routes[i].Domain.Name,
				Guid: getSummaryApp.Routes[i].Domain.GUID,
			},
		}
		cmd.pluginAppModel.Routes = append(cmd.pluginAppModel.Routes, routeSummary)
	}
	if cmd.pluginAppModel.Routes == nil {
		cmd.pluginAppModel.Routes = []plugin_models.GetApp_RouteSummary{}
	}

	for i := range getSummaryApp.Services {
		serviceSummary := plugin_models.GetApp_ServiceSummary{
			Name: getSummaryApp.Services[i].Name,
			Guid: getSummaryApp.Services[i].GUID,
		}
		cmd.pluginAppModel.Services = append(cmd.pluginAppModel.Services, serviceSummary)
	}
	if cmd.pluginAppModel.Services == nil {
		cmd.pluginAppModel.Services = []plugin_models.GetApp_ServiceSummary{}
	}
}
