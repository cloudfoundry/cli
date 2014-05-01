package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
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

func (command ListApps) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: "List all apps in the target space",
		Usage:       "CF_NAME apps",
	}
}

func (cmd ListApps) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ListApps) Run(c *cli.Context) {
	cmd.ui.Say("Getting apps in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apps, apiErr := cmd.appSummaryRepo.GetSummariesInCurrentSpace()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(apps) == 0 {
		cmd.ui.Say("No apps found")
		return
	}

	table := terminal.NewTable(cmd.ui, []string{
		"name",
		"requested state",
		"instances",
		"memory",
		"disk",
		"urls",
	})

	for _, application := range apps {
		var urls []string
		for _, route := range application.Routes {
			urls = append(urls, route.URL())
		}

		table.Add([]string{
			application.Name,
			ui_helpers.ColoredAppState(application.ApplicationFields),
			ui_helpers.ColoredAppInstances(application.ApplicationFields),
			formatters.ByteSize(application.Memory * formatters.MEGABYTE),
			formatters.ByteSize(application.DiskQuota * formatters.MEGABYTE),
			strings.Join(urls, ", "),
		})
	}

	table.Print()
}
