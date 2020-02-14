package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type SpaceUsersCommand struct {
	BaseCommand

	RequiredArgs    flag.SpaceUsersArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME space-users ORG SPACE"`
	relatedCommands interface{}         `related_commands:"org-users, orgs, set-space-role, spaces, unset-space-role"`
}

func (cmd *SpaceUsersCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting users in org {{.Org}} / space {{.Space}} as {{.CurrentUser}}...", map[string]interface{}{
		"Org":         cmd.RequiredArgs.Organization,
		"Space":       cmd.RequiredArgs.Space,
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.RequiredArgs.Space, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	spaceUsersByRoleType, warnings, err := cmd.Actor.GetSpaceUsersByRoleType(space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displaySpaceUsers(spaceUsersByRoleType)

	return nil
}

func (cmd SpaceUsersCommand) displaySpaceUsers(orgUsersByRoleType map[constant.RoleType][]v7action.User) {
	cmd.displayRoleGroup(orgUsersByRoleType[constant.SpaceManagerRole], "SPACE MANAGER")
	cmd.displayRoleGroup(orgUsersByRoleType[constant.SpaceDeveloperRole], "SPACE DEVELOPER")
	cmd.displayRoleGroup(orgUsersByRoleType[constant.SpaceAuditorRole], "SPACE AUDITOR")
}

func (cmd SpaceUsersCommand) displayRoleGroup(usersWithRole []v7action.User, roleLabel string) {
	v7action.SortUsers(usersWithRole)

	cmd.UI.DisplayHeader(roleLabel)
	if len(usersWithRole) > 0 {
		for _, userWithRole := range usersWithRole {
			cmd.UI.DisplayText("  {{.PresentationName}} ({{.Origin}})", map[string]interface{}{
				"PresentationName": userWithRole.PresentationName,
				"Origin":           v7action.GetHumanReadableOrigin(userWithRole),
			})
		}
	} else {
		cmd.UI.DisplayText("  No {{.RoleLabel}} found", map[string]interface{}{
			"RoleLabel": roleLabel,
		})
	}

	cmd.UI.DisplayNewline()
}
