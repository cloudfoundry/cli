package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
)

type BindSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.BindSecurityGroupV7Args `positional-args:"yes"`
	Lifecycle       flag.SecurityGroupLifecycle  `long:"lifecycle" choice:"running" choice:"staging" default:"running" description:"Lifecycle phase the group applies to."`
	Space           string                       `long:"space" description:"Space to bind the security group to. (Default: all existing spaces in org)"`
	usage           interface{}                  `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [--lifecycle (running | staging)] [--space SPACE]\n\nTIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}                  `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`
}

func (cmd BindSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
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

	spacesToBind := []resources.Space{}
	if cmd.Space != "" {
		var space resources.Space
		space, warnings, err = cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, space)
	} else {
		var spaces []resources.Space
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
	} else {

		if cmd.Space == "" {
			cmd.announceBinding(spacesToBind, user)
		}

		warnings, err = cmd.Actor.BindSecurityGroupToSpaces(securityGroup.GUID, spacesToBind, constant.SecurityGroupLifecycle(cmd.Lifecycle))
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()

		cmd.UI.DisplayText("TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.")
	}

	return nil
}

func (cmd BindSecurityGroupCommand) announceBinding(spaces []resources.Space, user configv3.User) {

	var spacenames []string

	for _, space := range spaces {
		spacenames = append(spacenames, space.Name)
	}

	tokens := map[string]interface{}{
		"lifecycle":      constant.SecurityGroupLifecycle(cmd.Lifecycle),
		"security_group": cmd.RequiredArgs.SecurityGroupName,
		"spaces":         strings.Join(spacenames, ", "),
		"organization":   cmd.RequiredArgs.OrganizationName,
		"username":       user.Name,
	}
	singular := "Assigning {{.lifecycle}} security group {{.security_group}} to space {{.spaces}} in org {{.organization}} as {{.username}}..."
	plural := "Assigning {{.lifecycle}} security group {{.security_group}} to spaces {{.spaces}} in org {{.organization}} as {{.username}}..."

	if len(spaces) == 1 {
		cmd.UI.DisplayTextWithFlavor(singular, tokens)
	} else {
		cmd.UI.DisplayTextWithFlavor(plural, tokens)
	}
}
