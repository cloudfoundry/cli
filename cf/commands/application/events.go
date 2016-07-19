package application

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/appevents"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Events struct {
	ui         terminal.UI
	config     coreconfig.Reader
	appReq     requirements.ApplicationRequirement
	eventsRepo appevents.Repository
}

func init() {
	commandregistry.Register(&Events{})
}

func (cmd *Events) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "events",
		Description: T("Show recent app events"),
		Usage: []string{
			"CF_NAME events ",
			T("APP_NAME"),
		},
	}
}

func (cmd *Events) Requirements(requirementsFactory requirements.Factory, c flags.FlagContext) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("events"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(c.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *Events) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.eventsRepo = deps.RepoLocator.GetAppEventsRepository()
	return cmd
}

func (cmd *Events) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("Getting events for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("time"), T("event"), T("actor"), T("description")})

	events, err := cmd.eventsRepo.RecentEvents(app.GUID, 50)
	if err != nil {
		return errors.New(T("Failed fetching events.\n{{.APIErr}}",
			map[string]interface{}{"APIErr": err.Error()}))
	}

	for _, event := range events {
		actor := event.ActorName
		if actor == "" {
			actor = event.Actor
		}

		table.Add(
			event.Timestamp.Local().Format("2006-01-02T15:04:05.00-0700"),
			event.Name,
			actor,
			event.Description,
		)
	}

	err = table.Print()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		cmd.ui.Say(T("No events for app {{.AppName}}",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))
		return nil
	}
	return nil
}
