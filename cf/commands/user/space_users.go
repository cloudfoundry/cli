package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"
)

var spaceRoles = []string{models.SPACE_MANAGER, models.SPACE_DEVELOPER, models.SPACE_AUDITOR}

type SpaceUsers struct {
	ui          terminal.UI
	config      core_config.Reader
	spaceRepo   spaces.SpaceRepository
	userRepo    api.UserRepository
	orgReq      requirements.OrganizationRequirement
	pluginModel *[]plugin_models.User
	pluginCall  bool
}

func init() {
	command_registry.Register(&SpaceUsers{})
}

func NewSpaceUsers(ui terminal.UI, config core_config.Reader, spaceRepo spaces.SpaceRepository, userRepo api.UserRepository) (cmd *SpaceUsers) {
	cmd = &SpaceUsers{}
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *SpaceUsers) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "space-users",
		Description: T("Show space users by role"),
		Usage:       T("CF_NAME space-users ORG SPACE"),
	}
}

func (cmd *SpaceUsers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + command_registry.Commands.CommandUsage("space-users"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *SpaceUsers) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Users

	return cmd
}

func (cmd *SpaceUsers) Execute(c flags.FlagContext) {
	spaceName := c.Args()[1]
	org := cmd.orgReq.GetOrganization()

	space, apiErr := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	var spaceRoleToDisplayName = map[string]string{
		models.SPACE_MANAGER:   T("SPACE MANAGER"),
		models.SPACE_DEVELOPER: T("SPACE DEVELOPER"),
		models.SPACE_AUDITOR:   T("SPACE AUDITOR"),
	}

	var usersMap = make(map[string]plugin_models.User)

	var users []models.UserFields
	for _, role := range spaceRoles {
		displayName := spaceRoleToDisplayName[role]

		if cmd.config.IsMinApiVersion("2.21.0") {
			users, apiErr = cmd.userRepo.ListUsersInSpaceForRoleWithNoUAA(space.Guid, role)
		} else {
			users, apiErr = cmd.userRepo.ListUsersInSpaceForRole(space.Guid, role)
		}

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)

			if cmd.pluginCall {
				u, found := usersMap[user.Username]
				if !found {
					u = plugin_models.User{}
					u.Username = user.Username
					u.Guid = user.Guid
					u.IsAdmin = user.IsAdmin
					u.Roles = make([]string, 1)
					u.Roles[0] = role
					usersMap[user.Username] = u
				} else {
					u.Roles = append(u.Roles, role)
					usersMap[user.Username] = u
				}
			}
		}

		if apiErr != nil {
			cmd.ui.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                  apiErr.Error(),
					"SpaceRoleToDisplayName": displayName,
				}))
			return
		}
	}

	if cmd.pluginCall {
		for _, v := range usersMap {
			*(cmd.pluginModel) = append(*(cmd.pluginModel), v)
		}
	}
}
