package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SSHEnabledCommand struct {
	command.BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME ssh-enabled APP_NAME"`
	relatedCommands interface{}  `related_commands:"enable-ssh, space-ssh-allowed, ssh"`
}

func (cmd *SSHEnabledCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	ccv3SSHEnabled, warnings, err := cmd.Actor.GetSSHEnabledByAppName(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if ccv3SSHEnabled.Enabled {
		cmd.UI.DisplayTextWithFlavor("ssh support is enabled for app '{{.AppName}}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("ssh support is disabled for app '{{.AppName}}'.", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})
		cmd.UI.DisplayText(ccv3SSHEnabled.Reason)
	}

	return nil
}
