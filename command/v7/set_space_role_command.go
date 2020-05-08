package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetSpaceRoleCommand struct {
	command.BaseCommand

	Args            flag.SpaceRoleArgs `positional-args:"yes"`
	IsClient        bool               `long:"client" description:"Assign a space role to a client-id of a (non-user) service account"`
	Origin          string             `long:"origin" description:"Indicates the identity provider to be used for authentication"`
	usage           interface{}        `usage:"CF_NAME set-space-role USERNAME ORG SPACE ROLE\n   CF_NAME set-space-role USERNAME ORG SPACE ROLE [--client]\n   CF_NAME set-space-role USERNAME ORG SPACE ROLE [--origin ORIGIN]\n\nROLES:\n   SpaceManager - Invite and manage users, and enable features for a given space\n   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n   SpaceAuditor - View logs, reports, and settings on this space"`
	relatedCommands interface{}        `related_commands:"space-users, unset-space-role"`
}

func (cmd *SetSpaceRoleCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Assigning role {{.RoleType}} to user {{.TargetUserName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUserName}}...", map[string]interface{}{
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
	if cmd.Origin == "" {
		origin = constant.DefaultOriginUaa
	}

	warnings, err = cmd.Actor.CreateSpaceRole(roleType, org.GUID, space.GUID, cmd.Args.Username, origin, cmd.IsClient)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.RoleAlreadyExistsError); ok {
			cmd.UI.DisplayWarning("User '{{.TargetUserName}}' already has role '{{.RoleType}}' in org '{{.OrgName}}' / space '{{.SpaceName}}'.", map[string]interface{}{
				"RoleType":       cmd.Args.Role.Role,
				"TargetUserName": cmd.Args.Username,
				"OrgName":        cmd.Args.Organization,
				"SpaceName":      cmd.Args.Space,
			})
		} else {
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd SetSpaceRoleCommand) validateFlags() error {
	if cmd.IsClient && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client", "--origin"},
		}
	}
	return nil
}

func convertSpaceRoleType(givenRole flag.SpaceRole) (constant.RoleType, error) {
	switch givenRole.Role {
	case "SpaceAuditor":
		return constant.SpaceAuditorRole, nil
	case "SpaceManager":
		return constant.SpaceManagerRole, nil
	case "SpaceDeveloper":
		return constant.SpaceDeveloperRole, nil
	default:
		return "", errors.New("Invalid role type.")
	}
}
