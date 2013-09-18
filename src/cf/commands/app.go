package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type App struct {
	ui             term.UI
	appSummaryRepo api.AppSummaryRepository
	appReq         requirements.ApplicationRequirement
}

func NewApp(ui term.UI, appSummaryRepo api.AppSummaryRepository) (cmd *App) {
	cmd = new(App)
	cmd.ui = ui
	cmd.appSummaryRepo = appSummaryRepo
	return
}

func (cmd *App) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
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

func (cmd *App) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ui.Say("Showing health and status for app %s...", term.EntityNameColor(app.Name))

	summary, err := cmd.appSummaryRepo.GetSummary(app)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n%s %s", term.HeaderColor("health:"), coloredState(summary.App.Health()))
	cmd.ui.Say("%s %s x %d instances", term.HeaderColor("usage:"), byteSize(summary.App.Memory*MEGABYTE), summary.App.Instances)
	cmd.ui.Say("%s %s\n", term.HeaderColor("urls:"), strings.Join(summary.App.Urls, ", "))

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

func (cmd *App) coloringFunc(value string, row int, col int) string {
	if row > 0 && col == 1 {
		return coloredState(value)
	}

	return term.DefaultColoringFunc(value, row, col)
}
