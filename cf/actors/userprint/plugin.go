package userprint

import (
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
)

type pluginPrinter struct {
	roles      []string
	userLister func(spaceGuid string, role string) ([]models.UserFields, error)
	users      userCollection
	printer    func([]userWithRoles)
}

type userCollection map[string]userWithRoles

type userWithRoles struct {
	models.UserFields
	Roles []string
}

func NewOrgUsersPluginPrinter(
	pluginModel *[]plugin_models.GetOrgUsers_Model,
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) *pluginPrinter {
	return &pluginPrinter{
		users:      make(userCollection),
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
			*pluginModel = orgUsers
		},
	}
}

func NewSpaceUsersPluginPrinter(
	pluginModel *[]plugin_models.GetSpaceUsers_Model,
	userLister func(guid string, role string) ([]models.UserFields, error),
	roles []string,
) *pluginPrinter {
	return &pluginPrinter{
		users:      make(userCollection),
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
			*pluginModel = spaceUsers
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

func (coll userCollection) storeAppendingRole(role string, username string, guid string, isAdmin bool) {
	u := coll[username]
	u.Roles = append(u.Roles, role)
	u.Username = username
	u.Guid = guid
	u.IsAdmin = isAdmin
	coll[username] = u
}

func (coll userCollection) all() (output []userWithRoles) {
	for _, u := range coll {
		output = append(output, u)
	}
	return output
}
