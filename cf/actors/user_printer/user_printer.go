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
	users       users
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
	users       users
}

type OrgUsersUiPrinter struct {
	UserPrinter
	Roles            []string
	RoleDisplayNames map[string]string
	UserLister       func(orgGuid string, role string) ([]models.UserFields, error)
	Ui               terminal.UI
}

type userWithRoles struct {
	models.UserFields
	Roles []string
}

type users struct {
	db map[string]userWithRoles
}

func NewOrgUsersPluginPrinter(
	pluginModel *[]plugin_models.GetOrgUsers_Model,
	userLister func(orgGuid string, role string) ([]models.UserFields, error),
	roles []string,
) (printer *OrgUsersPluginPrinter) {
	return &OrgUsersPluginPrinter{
		PluginModel: pluginModel,
		users:       users{db: make(map[string]userWithRoles)},
		UserLister:  userLister,
		Roles:       roles,
	}
}

func NewSpaceUsersPluginPrinter(
	pluginModel *[]plugin_models.GetSpaceUsers_Model,
	userLister func(orgGuid string, role string) ([]models.UserFields, error),
	roles []string,
) (printer *SpaceUsersPluginPrinter) {
	return &SpaceUsersPluginPrinter{
		PluginModel: pluginModel,
		users:       users{db: make(map[string]userWithRoles)},
		UserLister:  userLister,
		Roles:       roles,
	}
}

func (p *OrgUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(guid, role)
		for _, user := range users {
			p.users.StoreAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.PluginModel) = p.users.AsOrgUsers()
}

func (p *SpaceUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		users, _ := p.UserLister(guid, role)
		for _, user := range users {
			p.users.StoreAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.PluginModel) = p.users.AsSpaceUsers()
}

func (p *OrgUsersUiPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.Roles {
		displayName := p.RoleDisplayNames[role]
		users, err := p.UserLister(guid, role)
		if err != nil {
			p.Ui.Failed(T("Failed fetching org-users for role {{.OrgRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                err.Error(),
					"OrgRoleToDisplayName": displayName,
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

func (c *users) StoreAppendingRole(role string, username string, guid string, isAdmin bool) {
	u := c.db[username]
	u.Roles = append(u.Roles, role)
	u.Username = username
	u.Guid = guid
	u.IsAdmin = isAdmin
	c.db[username] = u
}

func (c *users) AsOrgUsers() (coll []plugin_models.GetOrgUsers_Model) {
	for _, u := range c.db {
		coll = append(coll, plugin_models.GetOrgUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.Roles,
		})
	}
	return coll
}

func (c *users) AsSpaceUsers() (coll []plugin_models.GetSpaceUsers_Model) {
	for _, u := range c.db {
		coll = append(coll, plugin_models.GetSpaceUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.Roles,
		})
	}
	return coll
}
