package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableSSHCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"allow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (cmd *EnableSSHCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Enabling ssh support for app {{.AppName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"AppName":         cmd.RequiredArgs.AppName,
		"CurrentUserName": user.Name,
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

	if appFeature.Enabled {
		cmd.UI.DisplayTextWithFlavor("ssh support for app '{{.AppName}}' is already enabled.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
	}

	updateSSHWarnings, err := cmd.Actor.UpdateAppFeature(app, true, "ssh")
	if err != nil {
		return err
	}
	cmd.UI.DisplayWarnings(updateSSHWarnings)

	sshEnabled, getSSHEnabledWarnings, err := cmd.Actor.GetSSHEnabled(app.GUID)
	if err != nil {
		return err
	}
	cmd.UI.DisplayWarnings(getSSHEnabledWarnings)

	cmd.UI.DisplayOK()

	if !sshEnabled.Enabled {
		cmd.UI.DisplayText("TIP: Ensure ssh is also enabled on the space and global level.")
	}

	if sshEnabled.Enabled && !appFeature.Enabled {
		cmd.UI.DisplayText("TIP: An app restart is required for the change to take effect.")
	}

	if appFeature.Enabled {
		cmd.UI.DisplayText("TIP: An app restart may be required for the change to take effect.")
	}

	return nil
}
