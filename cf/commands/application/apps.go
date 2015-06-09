package application

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
)

type ListApps struct {
	ui             terminal.UI
	config         core_config.Reader
	appSummaryRepo api.AppSummaryRepository
}

func init() {
	command_registry.Register(&ListApps{})
}

func (cmd *ListApps) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "apps",
		ShortName:   "a",
		Description: T("List all apps in the target space"),
		Usage:       "CF_NAME apps",
	}
}

func (cmd *ListApps) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed("Incorrect Usage. No argument required\n\n" + command_registry.Commands.CommandUsage("apps"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *ListApps) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	return cmd
}

func (cmd *ListApps) Execute(c flags.FlagContext) {
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
