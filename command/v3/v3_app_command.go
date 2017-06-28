package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V2AppActor

type V2AppActor interface {
	GetApplicationRoutes(appGUID string) ([]v2action.Route, v2action.Warnings, error)
}

//go:generate counterfeiter . V3AppActor

type V3AppActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string) (v3action.ApplicationSummary, v3action.Warnings, error)
}

type V3AppCommand struct {
	usage   interface{} `usage:"CF_NAME v3-app -n APP_NAME"`
	AppName string      `short:"n" long:"name" description:"The application name to display" required:"true"`

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

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2)

	cmd.AppSummaryDisplayer = shared.AppSummaryDisplayer{
		UI:              cmd.UI,
		Config:          cmd.Config,
		Actor:           cmd.Actor,
		V2AppRouteActor: v2Actor,
		AppName:         cmd.AppName,
	}
	return nil
}

func (cmd V3AppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	return cmd.AppSummaryDisplayer.DisplayAppInfo()
}
