package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV2 "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . AppActor

type AppActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool, routeActor v7action.RouteActor) (v7action.ApplicationSummary, v7action.Warnings, error)
}

type AppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME [--guid]"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	RouteActor  v7action.RouteActor
	Actor       AppActor
}

func (cmd *AppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.RouteActor = v2action.NewActor(ccClientV2, uaaClientV2, config)
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClientV2)

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
	summary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, false, cmd.RouteActor)
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
