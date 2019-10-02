package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

//go:generate counterfeiter . UpdateSSHActor

type UpdateSSHActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	UpdateSSH(app v7action.Application, enabled bool) (v7action.Warnings, error)
}

type EnableSSHCommand struct{
	RequiredArgs flag.AppName `positional-args:"yes"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UpdateSSHActor
}

func (cmd *EnableSSHCommand) Execute (args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	app, _, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	if err != nil {
		if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
			cmd.UI.DisplayText("App '{{.AppName}}' not found.", map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
			//cmd.UI.DisplayError(err) //TODO: (MdL) This feels weird but it fails without it?
			return err
		}
		return err
	}
	warnings, err := cmd.Actor.UpdateSSH(app, true)
	if err != nil {
		return err
	}

	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Enabling ssh support for app {{.AppName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"AppName": cmd.RequiredArgs.AppName,
		"CurrentUserName": username,
	})
	cmd.UI.DisplayWarnings(warnings)
	cmd.UI.DisplayOK()
	return nil
}
