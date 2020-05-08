package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/ui"
)

type EventsCommand struct {
	command.BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME events APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, logs, map-route, unmap-route"`
}

func (cmd EventsCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Getting events for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   appName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	events, warnings, err := cmd.Actor.GetRecentEventsByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		cmd.UI.DisplayText("No events found.")
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("time"),
			cmd.UI.TranslateText("event"),
			cmd.UI.TranslateText("actor"),
			cmd.UI.TranslateText("description"),
		},
	}

	for _, event := range events {
		table = append(table, []string{
			event.Time.Local().Format("2006-01-02T15:04:05.00-0700"),
			event.Type,
			event.ActorName,
			event.Description,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
