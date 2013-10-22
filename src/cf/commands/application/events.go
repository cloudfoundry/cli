package application

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strconv"
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

	cmd.ui.Say("Getting events for %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	appEvents, apiStatus := cmd.eventsRepo.ListEvents(app)
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching events.\n%s", apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(appEvents) == 0 {
		cmd.ui.Say("There are no events available for %s at this time", terminal.EntityNameColor(app.Name))
		return
	}

	if len(appEvents) == 1 {
		cmd.ui.Say("Showing 1 of 1 events...\n")
	} else {
		cmd.ui.Say("Showing all %d events...\n", len(appEvents))
	}

	table := [][]string{
		[]string{"time", "instance", "description", "exit status"},
	}

	for i := len(appEvents) - 1; i >= 0; i-- {
		event := appEvents[i]
		table = append(table, []string{
			event.Timestamp.Local().Format(TIMESTAMP_FORMAT),
			strconv.Itoa(event.InstanceIndex),
			event.ExitDescription,
			strconv.Itoa(event.ExitStatus),
		})
	}

	cmd.ui.DisplayTable(table)

}
