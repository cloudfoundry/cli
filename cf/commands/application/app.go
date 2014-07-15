package application

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/codegangsta/cli"
)

type ShowApp struct {
	ui             terminal.UI
	config         configuration.Reader
	appSummaryRepo api.AppSummaryRepository
	appStatsRepo   api.AppStatsRepository
	appReq         requirements.ApplicationRequirement
}

type ApplicationDisplayer interface {
	ShowApp(app models.Application)
}

func NewShowApp(ui terminal.UI, config configuration.Reader, appSummaryRepo api.AppSummaryRepository, appStatsRepo api.AppStatsRepository) (cmd *ShowApp) {
	cmd = new(ShowApp)
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	cmd.appStatsRepo = appStatsRepo
	return
}

func (cmd *ShowApp) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "app",
		Description: T("Display health and status for app"),
		Usage:       T("CF_NAME app APP"),
	}
}

func (cmd *ShowApp) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *ShowApp) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ShowApp(app)
}

func (cmd *ShowApp) ShowApp(app models.Application) {

	cmd.ui.Say(T("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
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

	var stats []models.AppStatsFields
	stats, apiErr = cmd.appStatsRepo.GetStats(app.Guid)
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

	cmd.ui.Say("%s %s\n", terminal.HeaderColor(T("urls:")), strings.Join(urls, ", "))

	if appIsStopped {
		cmd.ui.Say(T("There are no running instances of this app."))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{"", T("state"), T("since"), T("cpu"), T("memory"), T("disk")})

	for index, stat := range stats {
		since, err := time.Parse("2006-01-02 15:04:05 +0000", stat.Stats.Usage.Time)
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}

		uptime := time.Duration(stat.Stats.Uptime) * time.Second

		since = since.Add(-uptime)

		table.Add(
			fmt.Sprintf("#%d", index),
			ui_helpers.ColoredInstanceState(stat.State),
			since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", stat.Stats.Usage.Cpu*100),
			fmt.Sprintf(T("{{.MemUsage}} of {{.MemQuota}}",
				map[string]interface{}{
					"MemUsage": formatters.ByteSize(stat.Stats.Usage.Mem),
					"MemQuota": formatters.ByteSize(stat.Stats.MemQuota)})),
			fmt.Sprintf(T("{{.DiskUsage}} of {{.DiskQuota}}",
				map[string]interface{}{
					"DiskUsage": formatters.ByteSize(stat.Stats.Usage.Disk),
					"DiskQuota": formatters.ByteSize(stat.Stats.DiskQuota)})),
		)
	}

	table.Print()
}
