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

	cmd.ui.Say("Getting events for app %s in org %s / space %s as %s...\n",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	eventChan, statusChan := cmd.eventsRepo.ListEvents(app.Guid)
	table := cmd.ui.Table([]string{"time", "instance", "description", "exit status"})
	noEvents := true

	for events := range eventChan {
		rows := [][]string{}
		for i := len(events) - 1; i >= 0; i-- {
			event := events[i]
			rows = append(rows, []string{
				event.Timestamp.Local().Format(TIMESTAMP_FORMAT),
				strconv.Itoa(event.InstanceIndex),
				event.ExitDescription,
				strconv.Itoa(event.ExitStatus),
			})
		}
		table.Print(rows)
		noEvents = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching events.\n%s", apiStatus.Message)
		return
	}
	if noEvents {
		cmd.ui.Say("No events for app %s", terminal.EntityNameColor(app.Name))
		return
	}
}
