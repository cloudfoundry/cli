package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type BindSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.BindSecurityGroupV7Args `positional-args:"yes"`
	Lifecycle       flag.SecurityGroupLifecycle  `long:"lifecycle" choice:"running" choice:"staging" default:"running" description:"Lifecycle phase the group applies to."`
	Space           string                       `long:"space" description:"Space to bind the security group to. (Default: all existing spaces in org)"`
	usage           interface{}                  `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [--lifecycle (running | staging)] [--space SPACE]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}                  `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`
}

func (cmd BindSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if cmd.Space == "" {
		cmd.UI.DisplayTextWithFlavor("Assigning {{.lifecycle}} security group {{.security_group}} to all spaces in org {{.organization}} as {{.username}}...", map[string]interface{}{
			"lifecycle":      constant.SecurityGroupLifecycle(cmd.Lifecycle),
			"security_group": cmd.RequiredArgs.SecurityGroupName,
			"organization":   cmd.RequiredArgs.OrganizationName,
			"username":       user.Name,
		})
		cmd.UI.DisplayNewline()
	} else {
		cmd.UI.DisplayTextWithFlavor("Assigning {{.lifecycle}} security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...", map[string]interface{}{
			"lifecycle":      constant.SecurityGroupLifecycle(cmd.Lifecycle),
			"security_group": cmd.RequiredArgs.SecurityGroupName,
			"space":          cmd.Space,
			"organization":   cmd.RequiredArgs.OrganizationName,
			"username":       user.Name,
		})
	}

	securityGroup, warnings, err := cmd.Actor.GetSecurityGroup(cmd.RequiredArgs.SecurityGroupName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	spacesToBind := []v7action.Space{}
	if cmd.Space != "" {
		var space v7action.Space
		space, warnings, err = cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, space)
	} else {
		var spaces []v7action.Space
		spaces, warnings, err = cmd.Actor.GetOrganizationSpaces(org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, spaces...)
	}

	if len(spacesToBind) == 0 {
		cmd.UI.DisplayText("No spaces in org {{.organization}}.", map[string]interface{}{
			"organization": org.Name,
		})
	}

	for _, space := range spacesToBind {
		if len(spacesToBind) != 1 {
			cmd.UI.DisplayTextWithFlavor("Assigning {{.lifecycle}} security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...", map[string]interface{}{
				"lifecycle":      constant.SecurityGroupLifecycle(cmd.Lifecycle),
				"security_group": securityGroup.Name,
				"space":          space.Name,
				"organization":   org.Name,
				"username":       user.Name,
			})
		}

		warnings, err = cmd.Actor.BindSecurityGroupToSpace(securityGroup.GUID, space.GUID, constant.SecurityGroupLifecycle(cmd.Lifecycle))
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
