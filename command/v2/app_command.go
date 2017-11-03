package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . AppActor

type AppActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (v2action.ApplicationSummary, v2action.Warnings, error)
}

type AppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       AppActor
}

func (cmd *AppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd AppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	if cmd.GUID {
		return cmd.displayAppGUID()
	}

	return cmd.displayAppSummary()
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

func (cmd AppCommand) displayAppSummary() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	appSummary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	shared.DisplayAppSummary(cmd.UI, appSummary, false)

	return nil
}
