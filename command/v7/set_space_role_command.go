package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . SetSpaceRoleActor

type SetSpaceRoleActor interface {
	CreateSpaceRole(roleType constant.RoleType, userGUID string, spaceGUID string) (v7action.Role, v7action.Warnings, error)
	CreateOrgRole(roleType constant.RoleType, userGUID string, orgGUID string) (v7action.Role, v7action.Warnings, error)
	GetOrganizationByName(name string) (v7action.Organization, v7action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v7action.Space, v7action.Warnings, error)
	GetUser(username, origin string) (v7action.User, error)
}

type SetSpaceRoleCommand struct {
	Args              flag.SetSpaceRoleArgs `positional-args:"yes"`
	ClientCredentials bool                  `long:"client" description:"Assign a space role to a client-id of a (non-user) service account"`
	Origin            string                `long:"origin" description:"Indicates the identity provider to be used for authentication"`
	usage             interface{}           `usage:"CF_NAME set-space-role USERNAME ORG SPACE ROLE\n   CF_NAME set-space-role USERNAME ORG SPACE ROLE [--client CLIENT]\n   CF_NAME set-space-role USERNAME ORG SPACE ROLE [--origin ORIGIN]\n\nROLES:\n   SpaceManager - Invite and manage users, and enable features for a given space\n   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n   SpaceAuditor - View logs, reports, and settings on this space"`
	relatedCommands   interface{}           `related_commands:"space-users"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetSpaceRoleActor
}

func (cmd *SetSpaceRoleCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
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

	origin := cmd.Origin
	if cmd.Origin == "" {
		origin = "uaa"
	}

	targetUserGUID := cmd.Args.Username
	if !cmd.ClientCredentials {
		targetUser, err := cmd.Actor.GetUser(cmd.Args.Username, origin)
		if err != nil {
			return err
		}
		targetUserGUID = targetUser.GUID
	}

	roleType, err := convertSpaceRoleType(cmd.Args.Role)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Args.Organization)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.Args.Space, org.GUID)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	_, warnings, err = cmd.Actor.CreateOrgRole(constant.OrgUserRole, targetUserGUID, org.GUID)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		if _, isIdempotentError := err.(ccerror.RoleAlreadyExistsError); !isIdempotentError {
			return err
		}
	}

	_, warnings, err = cmd.Actor.CreateSpaceRole(roleType, targetUserGUID, space.GUID)
	cmd.UI.DisplayWarningsV7(warnings)

	if err != nil {
		if _, ok := err.(ccerror.RoleAlreadyExistsError); ok {
			cmd.UI.DisplayWarningV7("User '{{.TargetUserName}}' already has role '{{.RoleType}}' in org '{{.OrgName}}' / space '{{.SpaceName}}'.", map[string]interface{}{
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
	if cmd.ClientCredentials && cmd.Origin != "" {
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
