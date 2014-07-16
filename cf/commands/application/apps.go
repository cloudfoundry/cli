package application

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/codegangsta/cli"
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

func (cmd ListApps) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: T("List all apps in the target space"),
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
	cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	apps, apiErr := cmd.appSummaryRepo.GetSummariesInCurrentSpace()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(apps) == 0 {
		cmd.ui.Say(T("No apps found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("requested state"), T("instances"), T("memory"), T("disk"), T("urls")})

	for _, application := range apps {
		var urls []string
		for _, route := range application.Routes {
			urls = append(urls, route.URL())
		}

		table.Add(
			application.Name,
			ui_helpers.ColoredAppState(application.ApplicationFields),
			ui_helpers.ColoredAppInstances(application.ApplicationFields),
			formatters.ByteSize(application.Memory*formatters.MEGABYTE),
			formatters.ByteSize(application.DiskQuota*formatters.MEGABYTE),
			strings.Join(urls, ", "),
		)
	}

	table.Print()
}
