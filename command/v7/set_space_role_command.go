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

//go:generate counterfeiter . SetOrgRoleActor

type SetSpaceRoleActor interface {
	CreateSpaceRole(roleType constant.RoleType, userGUID string, spaceGUID string) (v7action.Role, v7action.Warnings, error)
	GetSpaceByName(name string) (v7action.Organization, v7action.Warnings, error)
	GetUser(username, origin string) (v7action.User, error)
}

type SetSpaceRoleCommand struct {
	Args              flag.SetOrgRoleArgs `positional-args:"yes"`
	ClientCredentials bool                `long:"client" description:"Assign an org role to a client-id of a (non-user) service account"`
	Origin            string              `long:"origin" description:"Origin for mapping a user account to a user in an external identity provider"`
	usage             interface{}         `usage:"CF_NAME set-org-role USERNAME ORG ROLE [--client]\n   CF_NAME set-space-role USERNAME ORG SPACE ROLE [--origin]\n\nROLES:\n   SpaceManager - Invite and manage users, select and change plans, and set spending limits\n   SpaceAuditor - Read-only access to org info and reports"`
	relatedCommands   interface{}         `related_commands:"org-users, set-org-role"`

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

	cmd.UI.DisplayTextWithFlavor("Assigning role {{.RoleType}} to user {{.TargetUserName}} in org {{.OrgName}} as {{.CurrentUserName}}...", map[string]interface{}{
		"RoleType":        cmd.Args.Role.Role,
		"TargetUserName":  cmd.Args.Username,
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

	roleType, err := convertRoleType(cmd.Args.Role)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetSpaceByName(cmd.Args.Organization)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	_, warnings, err = cmd.Actor.CreateSpaceRole(roleType, targetUserGUID, org.GUID)
	cmd.UI.DisplayWarningsV7(warnings)

	if err != nil {
		if _, ok := err.(ccerror.RoleAlreadyExistsError); ok {
			cmd.UI.DisplayWarningV7("User '{{.TargetUserName}}' already has role '{{.RoleType}}' in org '{{.OrgName}}'.", map[string]interface{}{
				"RoleType":       cmd.Args.Role.Role,
				"TargetUserName": cmd.Args.Username,
				"OrgName":        cmd.Args.Organization,
			})
		} else {
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd SetOrgRoleCommand) validateFlags() error {
	if cmd.ClientCredentials && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client", "--origin"},
		}
	}

	return nil
}

func convertRoleType(givenRole flag.OrgRole) (constant.RoleType, error) {
	switch givenRole.Role {
	case "OrgAuditor":
		return constant.OrgAuditorRole, nil
	case "OrgManager":
		return constant.OrgManagerRole, nil
	case "OrgBillingManager":
		return constant.OrgBillingManagerRole, nil
	default:
		return "", errors.New("Invalid role type.")
	}
}
