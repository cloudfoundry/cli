package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type App struct {
	ui             terminal.UI
	appSummaryRepo api.AppSummaryRepository
	appReq         requirements.ApplicationRequirement
}

func NewApp(ui terminal.UI, appSummaryRepo api.AppSummaryRepository) (cmd *App) {
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
	cmd.ui.Say("Showing health and status for app %s...", terminal.EntityNameColor(app.Name))

	summary, err := cmd.appSummaryRepo.GetSummary(app)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\nhealth: %s", summary.App.Health())
	cmd.ui.Say("usage: %dM x %d instances", summary.App.Memory, summary.App.Instances)
	cmd.ui.Say("urls: %s\n", strings.Join(summary.App.Urls, ", "))

	table := [][]string{
		[]string{"", "status", "since", "cpu", "memory", "disk"},
	}

	var byteSize = func(bytes int) string {
		unit := ""
		value := float64(bytes)

		const (
			byte  = 1.0
			kByte = 1024 * byte
			mByte = 1024 * kByte
			gByte = 1024 * mByte
			tByte = 1024 * gByte
		)

		switch {
		case bytes >= tByte:
			unit = "T"
			value = value / tByte
		case bytes >= gByte:
			unit = "G"
			value = value / gByte
		case bytes >= mByte:
			unit = "M"
			value = value / mByte
		case bytes >= kByte:
			unit = "K"
			value = value / kByte
		}

		stringValue := fmt.Sprintf("%.1f", value)
		stringValue = strings.TrimRight(stringValue, ".0")
		return fmt.Sprintf("%s%s", stringValue, unit)
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

	cmd.ui.DisplayTable(table, nil)
}
