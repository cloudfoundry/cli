package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type ShowApp struct {
	ui               terminal.UI
	config           configuration.Reader
	appSummaryRepo   api.AppSummaryRepository
	appInstancesRepo api.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
}

type ApplicationDisplayer interface {
	ShowApp(app models.Application)
}

func NewShowApp(ui terminal.UI, config configuration.Reader, appSummaryRepo api.AppSummaryRepository, appInstancesRepo api.AppInstancesRepository) (cmd *ShowApp) {
	cmd = new(ShowApp)
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	cmd.appInstancesRepo = appInstancesRepo
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
	cmd.ShowApp(app)
}

func (cmd *ShowApp) ShowApp(app models.Application) {

	cmd.ui.Say("Showing health and status for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	appSummary, apiResponse := cmd.appSummaryRepo.GetSummary(app.Guid)
	appIsStopped := (apiResponse != nil && (apiResponse.ErrorCode() == cf.APP_STOPPED ||
		apiResponse.ErrorCode() == cf.APP_NOT_STAGED)) ||
		appSummary.State == "stopped"

	if apiResponse != nil && !appIsStopped {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	var instances []models.AppInstanceFields
	instances, apiResponse = cmd.appInstancesRepo.GetInstances(app.Guid)
	if apiResponse != nil && !appIsStopped {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", terminal.HeaderColor("requested state:"), coloredAppState(appSummary.ApplicationFields))
	cmd.ui.Say("%s %s", terminal.HeaderColor("instances:"), coloredAppInstances(appSummary.ApplicationFields))
	cmd.ui.Say("%s %s x %d instances", terminal.HeaderColor("usage:"), formatters.ByteSize(appSummary.Memory*formatters.MEGABYTE), appSummary.InstanceCount)

	var urls []string
	for _, route := range appSummary.RouteSummaries {
		urls = append(urls, route.URL())
	}

	cmd.ui.Say("%s %s\n", terminal.HeaderColor("urls:"), strings.Join(urls, ", "))

	if appIsStopped {
		cmd.ui.Say("There are no running instances of this app.")
		return
	}

	table := [][]string{
		[]string{"", "state", "since", "cpu", "memory", "disk"},
	}

	for index, instance := range instances {
		table = append(table, []string{
			fmt.Sprintf("#%d", index),
			coloredInstanceState(instance),
			instance.Since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", instance.CpuUsage*100),
			fmt.Sprintf("%s of %s", formatters.ByteSize(instance.MemUsage), formatters.ByteSize(instance.MemQuota)),
			fmt.Sprintf("%s of %s", formatters.ByteSize(instance.DiskUsage), formatters.ByteSize(instance.DiskQuota)),
		})
	}

	cmd.ui.DisplayTable(table)
}
