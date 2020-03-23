package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type DisableSSHCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"disallow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (cmd *DisableSSHCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Disabling ssh support for app {{.AppName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"AppName":         cmd.RequiredArgs.AppName,
		"CurrentUserName": username,
	})

	app, getAppWarnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	if err != nil {
		return err
	}
	cmd.UI.DisplayWarnings(getAppWarnings)

	appFeature, getAppFeatureWarnings, err := cmd.Actor.GetAppFeature(app.GUID, "ssh")
	if err != nil {
		return err
	}

	cmd.UI.DisplayWarnings(getAppFeatureWarnings)

	if !appFeature.Enabled {
		cmd.UI.DisplayTextWithFlavor("ssh support for app '{{.AppName}}' is already disabled.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
	}

	updateSSHWarnings, err := cmd.Actor.UpdateAppFeature(app, false, "ssh")
	if err != nil {
		return err
	}

	cmd.UI.DisplayWarnings(updateSSHWarnings)
	cmd.UI.DisplayOK()
	return nil
}
