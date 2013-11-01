package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type ShowApp struct {
	ui             terminal.UI
	config         *configuration.Configuration
	appSummaryRepo api.AppSummaryRepository
	appReq         requirements.ApplicationRequirement
}

func NewShowApp(ui terminal.UI, config *configuration.Configuration, appSummaryRepo api.AppSummaryRepository) (cmd *ShowApp) {
	cmd = new(ShowApp)
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	return
}

func (cmd *ShowApp) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "app")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *ShowApp) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ui.Say("Showing health and status for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	summary, apiResponse := cmd.appSummaryRepo.GetSummary(app)
	appIsStopped := apiResponse.ErrorCode == cf.APP_STOPPED || apiResponse.ErrorCode == cf.APP_NOT_STAGED

	if apiResponse.IsNotSuccessful() && !appIsStopped {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", terminal.HeaderColor("state:"), coloredAppState(summary.App))
	cmd.ui.Say("%s %s", terminal.HeaderColor("instances:"), coloredAppInstaces(summary.App))
	cmd.ui.Say("%s %s x %d instances", terminal.HeaderColor("usage:"), formatters.ByteSize(summary.App.Memory*formatters.MEGABYTE), summary.App.Instances)

	var urls []string
	for _, route := range summary.App.Routes {
		urls = append(urls, route.URL())
	}
	cmd.ui.Say("%s %s\n", terminal.HeaderColor("urls:"), strings.Join(urls, ", "))

	if appIsStopped {
		return
	}

	table := [][]string{
		[]string{"", "status", "since", "cpu", "memory", "disk"},
	}

	for index, instance := range summary.Instances {
		table = append(table, []string{
			fmt.Sprintf("#%d", index),
			coloredInstanceState(instance),
			instance.Since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", instance.CpuUsage),
			fmt.Sprintf("%s of %s", formatters.ByteSize(instance.MemUsage), formatters.ByteSize(instance.MemQuota)),
			fmt.Sprintf("%s of %s", formatters.ByteSize(instance.DiskUsage), formatters.ByteSize(instance.DiskQuota)),
		})
	}

	cmd.ui.DisplayTable(table)
}
