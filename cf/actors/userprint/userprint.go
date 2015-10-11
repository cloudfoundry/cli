package userprint

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

type spaceUsersPluginPrinter struct {
	UserPrinter
	roles       []string
	userLister  func(spaceGuid string, role string) ([]models.UserFields, error)
	pluginModel *[]plugin_models.GetSpaceUsers_Model
	users       users
}

type orgUsersPluginPrinter struct {
	UserPrinter
	roles       []string
	userLister  func(orgGuid string, role string) ([]models.UserFields, error)
	pluginModel *[]plugin_models.GetOrgUsers_Model
	users       users
}

type SpaceUsersUiPrinter struct {
	UserPrinter
	Ui               terminal.UI
	UserLister       func(spaceGuid string, role string) ([]models.UserFields, error)
	Roles            []string
	RoleDisplayNames map[string]string
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
	roles []string
}

type users struct {
	db map[string]userWithRoles
}

func NewOrgUsersPluginPrinter(
	pluginModel *[]plugin_models.GetOrgUsers_Model,
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) (printer *orgUsersPluginPrinter) {
	return &orgUsersPluginPrinter{
		pluginModel: pluginModel,
		users:       users{db: make(map[string]userWithRoles)},
		userLister:  userLister,
		roles:       roles,
	}
}

func NewSpaceUsersPluginPrinter(
	pluginModel *[]plugin_models.GetSpaceUsers_Model,
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) (printer *spaceUsersPluginPrinter) {
	return &spaceUsersPluginPrinter{
		pluginModel: pluginModel,
		users:       users{db: make(map[string]userWithRoles)},
		userLister:  userLister,
		roles:       roles,
	}
}

func (p *orgUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.roles {
		users, _ := p.userLister(guid, role)
		for _, user := range users {
			p.users.storeAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.pluginModel) = p.users.asOrgUsers()
}

func (p *spaceUsersPluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.roles {
		users, _ := p.userLister(guid, role)
		for _, user := range users {
			p.users.storeAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	*(p.pluginModel) = p.users.asSpaceUsers()
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

func (c *users) storeAppendingRole(role string, username string, guid string, isAdmin bool) {
	u := c.db[username]
	u.roles = append(u.roles, role)
	u.Username = username
	u.Guid = guid
	u.IsAdmin = isAdmin
	c.db[username] = u
}

func (c *users) asOrgUsers() (coll []plugin_models.GetOrgUsers_Model) {
	for _, u := range c.db {
		coll = append(coll, plugin_models.GetOrgUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.roles,
		})
	}
	return coll
}

func (c *users) asSpaceUsers() (coll []plugin_models.GetSpaceUsers_Model) {
	for _, u := range c.db {
		coll = append(coll, plugin_models.GetSpaceUsers_Model{
			Guid:     u.Guid,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Roles:    u.roles,
		})
	}
	return coll
}
