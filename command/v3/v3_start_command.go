package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3StartActor

type V3StartActor interface {
	StartApplication(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
}

type V3StartCommand struct {
	usage   interface{} `usage:"CF_NAME v3-start -n APP_NAME"`
	AppName string      `short:"n" long:"name" description:"The application name to start" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3StartActor
}

func (cmd *V3StartCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config)

	return nil
}

func (cmd V3StartCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	_, warnings, err := cmd.Actor.StartApplication(cmd.AppName, cmd.Config.TargetedSpace().GUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
