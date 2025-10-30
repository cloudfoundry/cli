package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/flag"
)

type UnbindSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.UnbindSecurityGroupV7Args `positional-args:"yes"`
	Lifecycle       flag.SecurityGroupLifecycle    `long:"lifecycle" choice:"running" choice:"staging" default:"running" description:"Lifecycle phase the group applies to"`
	usage           interface{}                    `usage:"CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE [--lifecycle (running | staging)]\n\nTIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}                    `related_commands:"apps, restart, security-groups"`
}

func (cmd UnbindSecurityGroupCommand) Execute(args []string) error {
	var (
		err       error
		warnings  v7action.Warnings
		orgName   string
		spaceName string
	)

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	orgName = cmd.RequiredArgs.OrganizationName
	spaceName = cmd.RequiredArgs.SpaceName

	cmd.UI.DisplayTextWithFlavor("Unbinding {{.Lifecycle}} security group {{.SecurityGroupName}} from org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"Lifecycle":         string(cmd.Lifecycle),
		"SecurityGroupName": cmd.RequiredArgs.SecurityGroupName,
		"OrgName":           orgName,
		"SpaceName":         spaceName,
		"Username":          user.Name,
	})

	warnings, err = cmd.Actor.UnbindSecurityGroup(cmd.RequiredArgs.SecurityGroupName, orgName, spaceName, constant.SecurityGroupLifecycle(cmd.Lifecycle))
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, isNotBoundError := err.(actionerror.SecurityGroupNotBoundToSpaceError); isNotBoundError {
			cmd.UI.DisplayWarning(err.Error())

			cmd.UI.DisplayOK()
			cmd.UI.DisplayText("TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.")
			return nil
		}

		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
