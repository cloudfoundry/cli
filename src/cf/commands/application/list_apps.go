package application

import (
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ListApps struct {
	ui             terminal.UI
	config         *configuration.Configuration
	appSummaryRepo api.AppSummaryRepository
}

func NewListApps(ui terminal.UI, config *configuration.Configuration, appSummaryRepo api.AppSummaryRepository) (cmd ListApps) {
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	return
}

func (cmd ListApps) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ListApps) Run(c *cli.Context) {
	cmd.ui.Say("Getting apps in org %s and space %s as %s...",
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apps, apiResponse := cmd.appSummaryRepo.GetSummariesInCurrentSpace()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "state", "instances", "memory", "disk", "urls"},
	}

	for _, app := range apps {
		table = append(table, []string{
			app.Name,
			coloredAppState(app),
			coloredAppInstaces(app),
			formatters.ByteSize(app.Memory * formatters.MEGABYTE),
			formatters.ByteSize(app.DiskQuota * formatters.MEGABYTE),
			strings.Join(app.Urls, ", "),
		})
	}

	cmd.ui.DisplayTable(table)
}
