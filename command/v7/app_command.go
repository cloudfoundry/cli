package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type AppCommand struct {
	command.BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME [--guid]"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`
}

func (cmd AppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	if cmd.GUID {
		return cmd.displayAppGUID()
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	summary, warnings, err := cmd.Actor.GetDetailedAppSummary(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appSummaryDisplayer.AppDisplay(summary, false)
	return nil
}

func (cmd AppCommand) displayAppGUID() error {
	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(app.GUID)
	return nil
}
