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

type pluginPrinter struct {
	roles      []string
	userLister func(spaceGuid string, role string) ([]models.UserFields, error)
	users      users
	printer    func([]userWithRoles)
}

type SpaceUsersUiPrinter struct {
	Ui               terminal.UI
	UserLister       func(spaceGuid string, role string) ([]models.UserFields, error)
	Roles            []string
	RoleDisplayNames map[string]string
}

type OrgUsersUiPrinter struct {
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
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) *pluginPrinter {
	return &pluginPrinter{
		users:      users{db: make(map[string]userWithRoles)},
		userLister: userLister,
		roles:      roles,
		printer: func(users []userWithRoles) {
			var orgUsers []plugin_models.GetOrgUsers_Model
			for _, user := range users {
				orgUsers = append(orgUsers, plugin_models.GetOrgUsers_Model{
					Guid:     user.Guid,
					Username: user.Username,
					IsAdmin:  user.IsAdmin,
					Roles:    user.Roles,
				})
			}
			*(pluginModel) = orgUsers
		},
	}
}

func NewSpaceUsersPluginPrinter(
	pluginModel *[]plugin_models.GetSpaceUsers_Model,
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) *pluginPrinter {
	return &pluginPrinter{
		users:      users{db: make(map[string]userWithRoles)},
		userLister: userLister,
		roles:      roles,
		printer: func(users []userWithRoles) {
			var spaceUsers []plugin_models.GetSpaceUsers_Model
			for _, user := range users {
				spaceUsers = append(spaceUsers, plugin_models.GetSpaceUsers_Model{
					Guid:     user.Guid,
					Username: user.Username,
					IsAdmin:  user.IsAdmin,
					Roles:    user.Roles,
				})
			}
			*(pluginModel) = spaceUsers
		},
	}
}

func (p *pluginPrinter) PrintUsers(guid string, username string) {
	for _, role := range p.roles {
		users, _ := p.userLister(guid, role)
		for _, user := range users {
			p.users.storeAppendingRole(role, user.Username, user.Guid, user.IsAdmin)
		}
	}
	p.printer(p.users.all())
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
	u.Roles = append(u.Roles, role)
	u.Username = username
	u.Guid = guid
	u.IsAdmin = isAdmin
	c.db[username] = u
}

func (c *users) all() (coll []userWithRoles) {
	for _, u := range c.db {
		coll = append(coll, u)
	}
	return coll
}
