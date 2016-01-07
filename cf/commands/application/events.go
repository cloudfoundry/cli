package application

import (
	"github.com/cloudfoundry/cli/cf/api/app_events"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Events struct {
	ui         terminal.UI
	config     core_config.Reader
	appReq     requirements.ApplicationRequirement
	eventsRepo app_events.AppEventsRepository
}

func init() {
	command_registry.Register(&Events{})
}

func (cmd *Events) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "events",
		Description: T("Show recent app events"),
		Usage:       T("CF_NAME events APP_NAME"),
	}
}

func (cmd *Events) Requirements(requirementsFactory requirements.Factory, c flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("events"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Events) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.eventsRepo = deps.RepoLocator.GetAppEventsRepository()
	return cmd
}

func (cmd *Events) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("Getting events for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("time"), T("event"), T("actor"), T("description")})

	events, apiErr := cmd.eventsRepo.RecentEvents(app.Guid, 50)
	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching events.\n{{.ApiErr}}",
			map[string]interface{}{"ApiErr": apiErr.Error()}))
		return
	}

	for _, event := range events {
		table.Add(
			event.Timestamp.Local().Format("2006-01-02T15:04:05.00-0700"),
			event.Name,
			event.ActorName,
			event.Description,
		)
	}

	table.Print()

	if len(events) == 0 {
		cmd.ui.Say(T("No events for app {{.AppName}}",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))
		return
	}
}
