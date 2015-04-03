package application

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/codegangsta/cli"
)

type ListApps struct {
	ui             terminal.UI
	config         core_config.Reader
	appSummaryRepo api.AppSummaryRepository
	spaceRepo      spaces.SpaceRepository
}

func NewListApps(ui terminal.UI, config core_config.Reader, appSummaryRepo api.AppSummaryRepository, spaceRepo spaces.SpaceRepository) (cmd ListApps) {
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListApps) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: T("List all apps in space of targeted organization"),
		Usage:       "CF_NAME apps",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("s", T("Space name (To see the apps in specified space of current organization)")+
				T("\n\nTIP:\n")+T("    By default it will list all apps of current targeted space if '-s' flag is not provided.")),
		},
	}
}

func (cmd ListApps) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		cmd.ui.FailWithUsage(c)
	}
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ListApps) Run(c *cli.Context) {
	spaceName := c.String("s")
	var apps []models.Application
	var apiErr error
	var space models.Space
	if spaceName != "" {
		space, apiErr = cmd.spaceRepo.FindByName(spaceName)
		if apiErr != nil {
			cmd.ui.Failed(fmt.Sprintf(T("Unable to access space {{.SpaceName}}.\n{{.ApiErr}}",
				map[string]interface{}{"SpaceName": spaceName, "ApiErr": apiErr.Error()})))
			return
		}

		cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(space.Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))

		apps, apiErr = cmd.appSummaryRepo.GetSpaceSummaries(space.Guid)
	} else {
		cmd.ui.Say(T("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))

		apps, apiErr = cmd.appSummaryRepo.GetSummariesInCurrentSpace()
	}

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
