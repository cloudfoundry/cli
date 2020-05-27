package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindRunningSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-running-security-group SECURITY_GROUP\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}        `related_commands:"apps, restart, running-security-groups, security-groups"`
}

func (cmd UnbindRunningSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Unbinding security group {{.security_group}} from defaults for running as {{.username}}...", map[string]interface{}{
		"security_group": cmd.RequiredArgs.SecurityGroup,
		"username":       user.Name,
	})

	warnings, err := cmd.Actor.UpdateSecurityGroupGloballyEnabled(cmd.RequiredArgs.SecurityGroup, constant.SecurityGroupLifecycleRunning, false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, isGroupNotFoundError := err.(actionerror.SecurityGroupNotFoundError); isGroupNotFoundError {
			cmd.UI.DisplayWarning(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}

		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
