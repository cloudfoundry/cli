package userprint

import (
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SpaceUsersUIPrinter struct {
	UI               terminal.UI
	UserLister       func(spaceGUID string, role models.Role) ([]models.UserFields, error)
	Roles            []models.Role
	RoleDisplayNames map[models.Role]string
}

type OrgUsersUIPrinter struct {
	Roles            []models.Role
	RoleDisplayNames map[models.Role]string
	UserLister       func(orgGUID string, role models.Role) ([]models.UserFields, error)
	UI               terminal.UI
}

func (p *OrgUsersUIPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, err := p.UserLister(guid, role)
		if err != nil {
			p.UI.Failed(T("Failed fetching org-users for role {{.OrgRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                err.Error(),
					"OrgRoleToDisplayName": displayName,
				}))
			return
		}
		p.UI.Say("")
		p.UI.Say("%s", terminal.HeaderColor(displayName))

		if len(users) == 0 {
			p.UI.Say("  " + T("No {{.Role}} found", map[string]interface{}{
				"Role": displayName,
			}))
		} else {
			for _, user := range users {
				p.UI.Say("  %s", user.Username)
			}
		}
	}
}

func (p *SpaceUsersUIPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, err := p.UserLister(guid, role)
		if err != nil {
			p.UI.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                  err.Error(),
					"SpaceRoleToDisplayName": displayName,
				}))
			return
		}
		p.UI.Say("")
		p.UI.Say("%s", terminal.HeaderColor(displayName))

		if len(users) == 0 {
			p.UI.Say("  " + T("No {{.Role}} found", map[string]interface{}{
				"Role": displayName,
			}))
		} else {
			for _, user := range users {
				p.UI.Say("  %s", user.Username)
			}
		}
	}
}
