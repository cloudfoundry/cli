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
	config         configuration.Reader
	appSummaryRepo api.AppSummaryRepository
}

func NewListApps(ui terminal.UI, config configuration.Reader, appSummaryRepo api.AppSummaryRepository) (cmd ListApps) {
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
	cmd.ui.Say("Getting apps in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apps, apiResponse := cmd.appSummaryRepo.GetSummariesInCurrentSpace()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(apps) == 0 {
		cmd.ui.Say("No apps found")
		return
	}

	table := [][]string{
		[]string{"name", "requested state", "instances", "memory", "disk", "urls"},
	}

	for _, appSummary := range apps {
		var urls []string
		for _, route := range appSummary.RouteSummaries {
			urls = append(urls, route.URL())
		}

		table = append(table, []string{
			appSummary.Name,
			coloredAppState(appSummary.ApplicationFields),
			coloredAppInstances(appSummary.ApplicationFields),
			formatters.ByteSize(appSummary.Memory * formatters.MEGABYTE),
			formatters.ByteSize(appSummary.DiskQuota * formatters.MEGABYTE),
			strings.Join(urls, ", "),
		})
	}

	cmd.ui.DisplayTable(table)
}
