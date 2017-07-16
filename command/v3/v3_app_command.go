package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V2AppActor

type V2AppActor interface {
	GetApplicationRoutes(appGUID string) ([]v2action.Route, v2action.Warnings, error)
}

//go:generate counterfeiter . V3AppActor

type V3AppActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string) (v3action.ApplicationSummary, v3action.Warnings, error)
}

type V3AppCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	GUID         bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage        interface{}  `usage:"CF_NAME v3-app APP_NAME [--guid]"`

	UI                  command.UI
	Config              command.Config
	SharedActor         command.SharedActor
	Actor               V3AppActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
}

func (cmd *V3AppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config)

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)

	cmd.AppSummaryDisplayer = shared.AppSummaryDisplayer{
		UI:              cmd.UI,
		Config:          cmd.Config,
		Actor:           cmd.Actor,
		V2AppRouteActor: v2Actor,
		AppName:         cmd.RequiredArgs.AppName,
	}
	return nil
}

func (cmd V3AppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	if cmd.GUID {
		return cmd.displayAppGUID()
	}

	return cmd.AppSummaryDisplayer.DisplayAppInfo()
}

func (cmd V3AppCommand) displayAppGUID() error {
	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText(app.GUID)
	return nil
}
