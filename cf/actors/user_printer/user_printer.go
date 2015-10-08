package user_printer

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/models"
)

type UserPrinter interface {
	PrintUsers(guid string, username string)
}

type SpaceUsersPluginPrinter struct {
	UserPrinter
	Roles       []string
	UserLister  func(spaceGuid string, role string) ([]models.UserFields, error)
	PluginModel *[]plugin_models.GetSpaceUsers_Model
	Users       Users
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
	Roles       []string
	UserLister  func(orgGuid string, role string) ([]models.UserFields, error)
	PluginModel *[]plugin_models.GetOrgUsers_Model
	Users       Users
}

type OrgUsersUiPrinter struct {
	UserPrinter
	Roles            []string
	RoleDisplayNames map[string]string
	UserLister       func(orgGuid string, role string) ([]models.UserFields, error)
	Ui               terminal.UI
}

type UserWithRoles struct {
	models.UserFields
	Roles []string
}

type Users struct {
	Db map[string]UserWithRoles
}

func (p *OrgUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(guid, role)
		for _, user := range users {
			p.Users.StoreAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.PluginModel) = p.Users.AsOrgUsers()
}

func (p *SpaceUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(guid, role)
		for _, user := range users {
			p.Users.StoreAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.PluginModel) = p.Users.AsSpaceUsers()
}

func (p *OrgUsersUiPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, apiErr := p.UserLister(guid, role)

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

func (p *SpaceUsersUiPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, err := p.UserLister(guid, role)
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

func (c *Users) StoreAppendingRole(role string, username string, guid string, isAdmin bool) {
	u := c.Db[username]
	u.Roles = append(u.Roles, role)
	u.Username = username
	u.Guid = guid
	u.IsAdmin = isAdmin
	c.Db[username] = u
}

func (c *Users) AsOrgUsers() (coll []plugin_models.GetOrgUsers_Model) {
	for _, u := range c.Db {
		coll = append(coll, plugin_models.GetOrgUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.Roles,
		})
	}
	return coll
}

func (c *Users) AsSpaceUsers() (coll []plugin_models.GetSpaceUsers_Model) {
	for _, u := range c.Db {
		coll = append(coll, plugin_models.GetSpaceUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.Roles,
		})
	}
	return coll
}
