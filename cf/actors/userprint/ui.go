package userprint

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SpaceUsersUIPrinter struct {
	UI               terminal.UI
	UserLister       func(spaceGUID string, role string) ([]models.UserFields, error)
	Roles            []string
	RoleDisplayNames map[string]string
}

type OrgUsersUIPrinter struct {
	Roles            []string
	RoleDisplayNames map[string]string
	UserLister       func(orgGUID string, role string) ([]models.UserFields, error)
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
			p.UI.Say(fmt.Sprintf("  "+T("No %s found"), displayName))
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
			p.UI.Say(fmt.Sprintf("  "+T("No %s found"), displayName))
		} else {
			for _, user := range users {
				p.UI.Say("  %s", user.Username)
			}
		}
	}
}
