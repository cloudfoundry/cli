package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . V2V3AppSummaryActor

type V2V3AppSummaryActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (v2v3action.ApplicationSummary, v2v3action.Warnings, error)
}

//go:generate counterfeiter . V3AppActor

type V3AppActor interface {
	shared.V3AppSummaryActor
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
}

type V3AppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME [--guid]"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`

	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
	AppSummaryActor V2V3AppSummaryActor
	Actor           V3AppActor
}

func (cmd *V3AppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	ccClientV2, uaaClientV2, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)
	v3Actor := v3action.NewActor(ccClient, config, nil, nil)
	cmd.AppSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)
	cmd.Actor = v3Actor

	return nil
}

func (cmd V3AppCommand) Execute(args []string) error {
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

	appSummaryDisplayer := shared.NewAppSummaryDisplayer2(cmd.UI)
	summary, warnings, err := cmd.AppSummaryActor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appSummaryDisplayer.AppDisplay(summary, false)
	return nil
}

func (cmd V3AppCommand) displayAppGUID() error {
	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(app.GUID)
	return nil
}
