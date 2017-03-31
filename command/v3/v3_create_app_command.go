package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3CreateAppActor

type V3CreateAppActor interface {
	CreateApplicationByNameAndSpace(name string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
}

type V3CreateAppCommand struct {
	usage   interface{} `usage:"CF_NAME v3-create-app --name [name]"`
	AppName string      `short:"n" long:"name" description:"The desired application name" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3CreateAppActor
}

func (cmd *V3CreateAppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd V3CreateAppCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating V3 app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})

	_, warnings, err := cmd.Actor.CreateApplicationByNameAndSpace(cmd.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if _, ok := err.(v3action.ApplicationAlreadyExistsError); ok {
		cmd.UI.DisplayWarning("App {{.AppName}} already exists.", map[string]interface{}{
			"AppName": cmd.AppName,
		})
	} else if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
