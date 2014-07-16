package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Events struct {
	ui         terminal.UI
	config     configuration.Reader
	appReq     requirements.ApplicationRequirement
	eventsRepo api.AppEventsRepository
}

func NewEvents(ui terminal.UI, config configuration.Reader, eventsRepo api.AppEventsRepository) (cmd *Events) {
	cmd = new(Events)
	cmd.ui = ui
	cmd.config = config
	cmd.eventsRepo = eventsRepo
	return
}

func (cmd *Events) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "events",
		Description: T("Show recent app events"),
		Usage:       T("CF_NAME events APP"),
	}
}

func (cmd *Events) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Events) Run(c *cli.Context) {
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
