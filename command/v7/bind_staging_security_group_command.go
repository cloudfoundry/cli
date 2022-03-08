package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type BindStagingSecurityGroupCommand struct {
	BaseCommand

	SecurityGroup   flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME bind-staging-security-group SECURITY_GROUP\n\nTIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}        `related_commands:"apps, bind-running-security-group, bind-security-group, restart, security-groups, staging-security-groups"`
}

func (cmd BindStagingSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Binding security group {{.security_group}} to staging as {{.username}}...", map[string]interface{}{
		"security_group": cmd.SecurityGroup.SecurityGroup,
		"username":       user.Name,
	})

	warnings, err := cmd.Actor.UpdateSecurityGroupGloballyEnabled(cmd.SecurityGroup.SecurityGroup, constant.SecurityGroupLifecycleStaging, true)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayText("TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
