package userprint

import (
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/plugin/models"
)

type pluginPrinter struct {
	roles      []models.Role
	userLister func(spaceGUID string, role models.Role) ([]models.UserFields, error)
	users      userCollection
	printer    func([]userWithRoles)
}

type userCollection map[string]userWithRoles

type userWithRoles struct {
	models.UserFields
	Roles []models.Role
}

func NewOrgUsersPluginPrinter(
	pluginModel *[]plugin_models.GetOrgUsers_Model,
	userLister func(guid string, role models.Role) ([]models.UserFields, error),
	roles []models.Role,
) *pluginPrinter {
	return &pluginPrinter{
		users:      make(userCollection),
		userLister: userLister,
		roles:      roles,
		printer: func(users []userWithRoles) {
			var orgUsers []plugin_models.GetOrgUsers_Model
			for _, user := range users {
				orgUsers = append(orgUsers, plugin_models.GetOrgUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  user.IsAdmin,
					Roles:    rolesToString(user.Roles),
				})
			}
			*pluginModel = orgUsers
		},
	}
}

func NewSpaceUsersPluginPrinter(
	pluginModel *[]plugin_models.GetSpaceUsers_Model,
	userLister func(guid string, role models.Role) ([]models.UserFields, error),
	roles []models.Role,
) *pluginPrinter {
	return &pluginPrinter{
		users:      make(userCollection),
		userLister: userLister,
		roles:      roles,
		printer: func(users []userWithRoles) {
			var spaceUsers []plugin_models.GetSpaceUsers_Model
			for _, user := range users {
				spaceUsers = append(spaceUsers, plugin_models.GetSpaceUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  user.IsAdmin,
					Roles:    rolesToString(user.Roles),
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
			p.users.storeAppendingRole(role, user.Username, user.GUID, user.IsAdmin)
		}
	}
	p.printer(p.users.all())
}

func (coll userCollection) storeAppendingRole(role models.Role, username string, guid string, isAdmin bool) {
	u := coll[username]
	u.Roles = append(u.Roles, role)
	u.Username = username
	u.GUID = guid
	u.IsAdmin = isAdmin
	coll[username] = u
}

func (coll userCollection) all() (output []userWithRoles) {
	for _, u := range coll {
		output = append(output, u)
	}
	return output
}

func rolesToString(roles []models.Role) []string {
	var rolesStr []string
	for _, role := range roles {
		rolesStr = append(rolesStr, role.ToString())
	}
	return rolesStr
}
