package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Events struct {
	ui         terminal.UI
	config     *configuration.Configuration
	appReq     requirements.ApplicationRequirement
	eventsRepo api.AppEventsRepository
}

func NewEvents(ui terminal.UI, config *configuration.Configuration, eventsRepo api.AppEventsRepository) (cmd *Events) {
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
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	table := cmd.ui.Table([]string{"time", "event", "description"})
	noEvents := true

	apiResponse := cmd.eventsRepo.ListEvents(app.Guid, api.ListEventsCallback(func(events []cf.EventFields) (fetchNext bool) {
		rows := [][]string{}
		for _, event := range events {
			rows = append(rows, []string{
				event.Timestamp.Local().Format(TIMESTAMP_FORMAT),
				event.Name,
				event.Description,
			})
		}
		table.Print(rows)
		noEvents = false
		return true
	}))

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching events.\n%s", apiResponse.Message)
		return
	}
	if noEvents {
		cmd.ui.Say("No events for app %s", terminal.EntityNameColor(app.Name))
		return
	}
}
