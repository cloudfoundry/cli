package user_printer

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/models"
)

type UserPrinter interface {
	PrintUsers(org models.Organization, space models.Space, username string)
}

type SpaceUsersPluginPrinter struct {
	UserPrinter
	UsersMap    map[string]plugin_models.GetSpaceUsers_Model
	UserLister  func(spaceGuid string, role string) ([]models.UserFields, error)
	Roles       []string
	PluginModel *[]plugin_models.GetSpaceUsers_Model
}

type SpaceUsersUiPrinter struct {
	UserPrinter
	Ui               terminal.UI
	UserLister       func(spaceGuid string, role string) ([]models.UserFields, error)
	Roles            []string
	RoleDisplayNames map[string]string
}

type OrgUsersPluginPrinter struct {
	UserPrinter
	UsersMap    map[string]plugin_models.GetOrgUsers_Model
	Roles       []string
	UserLister  func(orgGuid string, role string) ([]models.UserFields, error)
	PluginModel *[]plugin_models.GetOrgUsers_Model
}

type OrgUsersUiPrinter struct {
	UserPrinter
	Roles            []string
	RoleDisplayNames map[string]string
	UserLister       func(orgGuid string, role string) ([]models.UserFields, error)
	Ui               terminal.UI
}

func (p *OrgUsersPluginPrinter) PrintUsers(org models.Organization, _ models.Space, username string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(org.Guid, role)
		for _, user := range users {
			u, found := p.UsersMap[user.Username]
			if found {
				u.Roles = append(u.Roles, role)
				p.UsersMap[user.Username] = u
			} else {
				u = plugin_models.GetOrgUsers_Model{}
				u.Username = user.Username
				u.Guid = user.Guid
				u.IsAdmin = user.IsAdmin
				u.Roles = make([]string, 1)
				u.Roles[0] = role
				p.UsersMap[user.Username] = u
			}
		}
	}
	for _, v := range p.UsersMap {
		*(p.PluginModel) = append(*(p.PluginModel), v)
	}
}

func (p *OrgUsersUiPrinter) PrintUsers(org models.Organization, _ models.Space, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, apiErr := p.UserLister(org.Guid, role)

		p.Ui.Say("")
		p.Ui.Say("%s", terminal.HeaderColor(displayName))

		if len(users) == 0 {
			p.Ui.Say(fmt.Sprintf("  "+T("No %s found"), displayName))
			continue
		}

		for _, user := range users {
			p.Ui.Say("  %s", user.Username)
		}

		if apiErr != nil {
			p.Ui.Failed(T("Failed fetching org-users for role {{.OrgRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                apiErr.Error(),
					"OrgRoleToDisplayName": displayName,
				}))
			return
		}
	}
}

func (p *SpaceUsersPluginPrinter) PrintUsers(_ models.Organization, space models.Space, _ string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(space.Guid, role)
		for _, user := range users {
			u, found := p.UsersMap[user.Username]
			if found {
				u.Roles = append(u.Roles, role)
			} else {
				u = plugin_models.GetSpaceUsers_Model{}
				u.Username = user.Username
				u.Guid = user.Guid
				u.IsAdmin = user.IsAdmin
				u.Roles = make([]string, 1)
				u.Roles[0] = role
			}
			p.UsersMap[user.Username] = u
		}
	}
	for _, v := range p.UsersMap {
		*(p.PluginModel) = append(*(p.PluginModel), v)
	}
}

func (p *SpaceUsersUiPrinter) PrintUsers(org models.Organization, space models.Space, username string) {
	p.Ui.Say(T("Getting users in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(username),
		}))

	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, err := p.UserLister(space.Guid, role)
		if err != nil {
			p.Ui.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                  err.Error(),
					"SpaceRoleToDisplayName": displayName,
				}))
			return
		}
		p.Ui.Say("")
		p.Ui.Say("%s", terminal.HeaderColor(displayName))

		if len(users) == 0 {
			p.Ui.Say(fmt.Sprintf("  "+T("No %s found"), displayName))
		} else {
			for _, user := range users {
				p.Ui.Say("  %s", user.Username)
			}
		}
	}
}
