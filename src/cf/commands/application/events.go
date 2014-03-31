package application

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
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

func (cmd *Events) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "events")
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

func (cmd *Events) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Getting events for app %s in org %s / space %s as %s...\n",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	table := cmd.ui.Table([]string{"time", "event", "description"})

	events, apiErr := cmd.eventsRepo.RecentEvents(app.Guid, 50)
	if apiErr != nil {
		cmd.ui.Failed("Failed fetching events.\n%s", apiErr.Error())
		return
	}

	for _, event := range events {
		table.Print([][]string{{
			event.Timestamp.Local().Format(TIMESTAMP_FORMAT),
			event.Name,
			event.Description,
		}})
	}

	if len(events) == 0 {
		cmd.ui.Say("No events for app %s", terminal.EntityNameColor(app.Name))
		return
	}
}
