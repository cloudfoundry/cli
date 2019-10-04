package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateAppActor

type CreateAppActor interface {
	CreateApplicationInSpace(app v7action.Application, spaceGUID string) (v7action.Application, v7action.Warnings, error)
}

type CreateAppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	AppType         flag.AppType `long:"app-type" choice:"buildpack" choice:"docker" description:"App lifecycle type to stage and run the app" default:"buildpack"`
	usage           interface{}  `usage:"CF_NAME create-app APP_NAME [--app-type (buildpack | docker)]"`
	relatedCommands interface{}  `related_commands:"app, apps, push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateAppActor
}

func (cmd *CreateAppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(client, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd CreateAppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})

	_, warnings, err := cmd.Actor.CreateApplicationInSpace(
		v7action.Application{
			Name:          cmd.RequiredArgs.AppName,
			LifecycleType: constant.AppLifecycleType(cmd.AppType),
		},
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case ccerror.NameNotUniqueInSpaceError:
			cmd.UI.DisplayText(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
