package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsetOrgRoleCommand struct {
	BaseCommand

	Args            flag.OrgRoleArgs `positional-args:"yes"`
	IsClient        bool             `long:"client" description:"Unassign an org role for a client-id of a (non-user) service account"`
	Origin          string           `long:"origin" description:"Indicates the identity provider to be used for authentication"`
	usage           interface{}      `usage:"CF_NAME unset-org-role USERNAME ORG ROLE\n   CF_NAME unset-org-role USERNAME ORG ROLE [--client]\n   CF_NAME unset-org-role USERNAME ORG ROLE [--origin ORIGIN]\n\nROLES:\n   OrgManager - Invite and manage users, select and change plans, and set spending limits\n   BillingManager - Create and manage the billing account and payment info\n   OrgAuditor - Read-only access to org info and reports"`
	relatedCommands interface{}      `related_commands:"org-users, set-space-role"`
}

func (cmd *UnsetOrgRoleCommand) Execute(args []string) error {
	err := cmd.validateFlags()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing role {{.RoleType}} from user {{.TargetUserName}} in org {{.OrgName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"RoleType":        cmd.Args.Role.Role,
		"TargetUserName":  cmd.Args.Username,
		"OrgName":         cmd.Args.Organization,
		"CurrentUserName": currentUser.Name,
	})

	roleType, err := convertRoleType(cmd.Args.Role)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Args.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	origin := cmd.Origin
	if origin == "" {
		origin = constant.DefaultOriginUaa
	}

	warnings, err = cmd.Actor.DeleteOrgRole(roleType, org.GUID, cmd.Args.Username, origin, cmd.IsClient)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd UnsetOrgRoleCommand) validateFlags() error {
	if cmd.IsClient && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client", "--origin"},
		}
	}
	return nil
}
