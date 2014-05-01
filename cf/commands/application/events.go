package application

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
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

func (command *Events) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "events",
		Description: "Show recent app events",
		Usage:       "CF_NAME events APP",
	}
}

func (cmd *Events) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "events")
		return
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

	cmd.ui.Say("Getting events for app %s in org %s / space %s as %s...\n",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	table := cmd.ui.Table([]string{"time", "event", "actor", "description"})

	events, apiErr := cmd.eventsRepo.RecentEvents(app.Guid, 50)
	if apiErr != nil {
		cmd.ui.Failed("Failed fetching events.\n%s", apiErr.Error())
		return
	}

	for _, event := range events {
		table.Add([]string{
			event.Timestamp.Local().Format("2006-01-02T15:04:05.00-0700"),
			event.Name,
			event.ActorName,
			event.Description,
		})
	}

	table.Print()

	if len(events) == 0 {
		cmd.ui.Say("No events for app %s", terminal.EntityNameColor(app.Name))
		return
	}
}
