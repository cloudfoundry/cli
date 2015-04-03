package application

import (
	"fmt"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/codegangsta/cli"
)

type ShowApp struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	appLogsNoaaRepo  api.LogsNoaaRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
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

func (cmd *ShowApp) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "app",
		Description: T("Display health and status for app"),
		Usage:       T("CF_NAME app APP_NAME"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given app's guid.  All other health and status output for the app is suppressed.")},
		},
	}
}

func (cmd *ShowApp) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *ShowApp) Run(c *cli.Context) {
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
		cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("stack:")), app.Stack.Name)
	} else {
		cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("stack:")), "unknown")
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
