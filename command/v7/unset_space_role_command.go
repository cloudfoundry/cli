package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsetSpaceRoleCommand struct {
	command.BaseCommand

	Args            flag.SpaceRoleArgs `positional-args:"yes"`
	IsClient        bool               `long:"client" description:"Remove space role from a client-id of a (non-user) service account"`
	Origin          string             `long:"origin" description:"Indicates the identity provider to be used for authentication"`
	usage           interface{}        `usage:"CF_NAME unset-space-role USERNAME ORG SPACE ROLE\n   CF_NAME unset-space-role USERNAME ORG SPACE ROLE [--client]\n   CF_NAME unset-space-role USERNAME ORG SPACE ROLE [--origin ORIGIN]\n\nROLES:\n   SpaceManager - Invite and manage users, and enable features for a given space\n   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n   SpaceAuditor - View logs, reports, and settings on this space"`
	relatedCommands interface{}        `related_commands:"set-space-role, space-users"`
}

func (cmd *UnsetSpaceRoleCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Removing role {{.RoleType}} from user {{.TargetUserName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"RoleType":        cmd.Args.Role.Role,
		"TargetUserName":  cmd.Args.Username,
		"OrgName":         cmd.Args.Organization,
		"SpaceName":       cmd.Args.Space,
		"CurrentUserName": currentUser.Name,
	})

	roleType, err := convertSpaceRoleType(cmd.Args.Role)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Args.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.Args.Space, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	origin := cmd.Origin
	if origin == "" {
		origin = constant.DefaultOriginUaa
	}

	warnings, err = cmd.Actor.DeleteSpaceRole(roleType, space.GUID, cmd.Args.Username, origin, cmd.IsClient)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd UnsetSpaceRoleCommand) validateFlags() error {
	if cmd.IsClient && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client", "--origin"},
		}
	}
	return nil
}
