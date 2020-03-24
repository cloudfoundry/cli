package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type BindStagingSecurityGroupCommand struct {
	BaseCommand

	SecurityGroupName string      `positional-args:"yes"`
	usage             interface{} `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [--lifecycle (running | staging)] [--space SPACE]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands   interface{} `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`
}

func (cmd BindStagingSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Binding security group {{.security_group}} to staging as {{.username}}...", map[string]interface{}{
		"security_group": cmd.SecurityGroupName,
		"username":       user.Name,
	})

	warnings, err := cmd.Actor.UpdateSecurityGroupGloballyEnabled(cmd.SecurityGroupName, constant.SecurityGroupLifecycleStaging, true)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
