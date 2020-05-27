package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
)

type OrgUsersCommand struct {
	BaseCommand

	RequiredArgs    flag.Organization `positional-args:"yes"`
	AllUsers        bool              `long:"all-users" short:"a" description:"List all users with roles in the org or in spaces within the org"`
	usage           interface{}       `usage:"CF_NAME org-users ORG"`
	relatedCommands interface{}       `related_commands:"orgs, set-org-role"`
}

func (cmd *OrgUsersCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting users in org {{.Org}} as {{.CurrentUser}}...", map[string]interface{}{
		"Org":         cmd.RequiredArgs.Organization,
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	orgUsersByRoleType, warnings, err := cmd.Actor.GetOrgUsersByRoleType(org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayOrgUsers(orgUsersByRoleType)

	return nil
}

func (cmd OrgUsersCommand) displayOrgUsers(orgUsersByRoleType map[constant.RoleType][]resources.User) {
	if cmd.AllUsers {
		cmd.displayRoleGroup(getUniqueUsers(orgUsersByRoleType), "ORG USERS")
	} else {
		cmd.displayRoleGroup(orgUsersByRoleType[constant.OrgManagerRole], "ORG MANAGER")
		cmd.displayRoleGroup(orgUsersByRoleType[constant.OrgBillingManagerRole], "BILLING MANAGER")
		cmd.displayRoleGroup(orgUsersByRoleType[constant.OrgAuditorRole], "ORG AUDITOR")
	}
}

func (cmd OrgUsersCommand) displayRoleGroup(usersWithRole []resources.User, roleLabel string) {
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

func getUniqueUsers(orgUsersByRoleType map[constant.RoleType][]resources.User) []resources.User {
	var allUsers []resources.User

	usersSet := make(map[string]bool)
	addUsersWithType := func(roleType constant.RoleType) {
		for _, user := range orgUsersByRoleType[roleType] {
			if _, ok := usersSet[user.GUID]; !ok {
				allUsers = append(allUsers, user)
			}

			usersSet[user.GUID] = true
		}
	}

	addUsersWithType(constant.OrgUserRole)
	addUsersWithType(constant.OrgManagerRole)
	addUsersWithType(constant.OrgBillingManagerRole)
	addUsersWithType(constant.OrgAuditorRole)

	return allUsers
}
