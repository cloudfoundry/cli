package application

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type ShowApp struct {
	ui             terminal.UI
	appSummaryRepo api.AppSummaryRepository
	appReq         requirements.ApplicationRequirement
}

func NewShowApp(ui terminal.UI, appSummaryRepo api.AppSummaryRepository) (cmd *ShowApp) {
	cmd = new(ShowApp)
	cmd.ui = ui
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
	cmd.ui.Say("Showing health and status for app %s...", terminal.EntityNameColor(app.Name))

	summary, apiResponse := cmd.appSummaryRepo.GetSummary(app)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", terminal.HeaderColor("health:"), coloredState(summary.App.Health()))
	cmd.ui.Say("%s %s x %d instances", terminal.HeaderColor("usage:"), byteSize(summary.App.Memory*MEGABYTE), summary.App.Instances)
	cmd.ui.Say("%s %s\n", terminal.HeaderColor("urls:"), strings.Join(summary.App.Urls, ", "))

	table := [][]string{
		[]string{"", "status", "since", "cpu", "memory", "disk"},
	}

	for index, instance := range summary.Instances {
		table = append(table, []string{
			fmt.Sprintf("#%d", index),
			string(instance.State),
			instance.Since.Format("2006-01-02 03:04:05 PM"),
			fmt.Sprintf("%.1f%%", instance.CpuUsage),
			fmt.Sprintf("%s of %s", byteSize(instance.MemUsage), byteSize(instance.MemQuota)),
			fmt.Sprintf("%s of %s", byteSize(instance.DiskUsage), byteSize(instance.DiskQuota)),
		})
	}

	cmd.ui.DisplayTable(table, cmd.coloringFunc)
}

func (cmd *ShowApp) coloringFunc(value string, row int, col int) string {
	if row > 0 && col == 1 {
		return coloredState(value)
	}

	return terminal.DefaultColoringFunc(value, row, col)
}
