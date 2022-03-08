package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindStagingSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-staging-security-group SECURITY_GROUP\n\nTIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}        `related_commands:"apps, restart, security-groups, staging-security-groups"`
}

func (cmd UnbindStagingSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Unbinding security group {{.security_group}} from defaults for staging as {{.username}}...", map[string]interface{}{
		"security_group": cmd.RequiredArgs.SecurityGroup,
		"username":       user.Name,
	})

	warnings, err := cmd.Actor.UpdateSecurityGroupGloballyEnabled(cmd.RequiredArgs.SecurityGroup, constant.SecurityGroupLifecycleStaging, false)
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
	cmd.UI.DisplayText("TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
