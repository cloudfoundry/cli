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

type SpaceUsers struct {
	ui          terminal.UI
	config      core_config.Reader
	spaceRepo   spaces.SpaceRepository
	userRepo    api.UserRepository
	orgReq      requirements.OrganizationRequirement
	pluginModel *[]plugin_models.GetSpaceUsers_Model
	pluginCall  bool
}

type userPrinter interface {
	printUsers()
}

type pluginPrinter struct {
	userPrinter
	usersMap               map[string]plugin_models.GetSpaceUsers_Model
	userLister             func(spaceGuid string, role string) ([]models.UserFields, error)
	spaceRoleToDisplayName map[string]string
	space                  models.Space
	pluginModel            *[]plugin_models.GetSpaceUsers_Model
}

type uiPrinter struct {
	userPrinter
	ui                     terminal.UI
	userLister             func(spaceGuid string, role string) ([]models.UserFields, error)
	spaceRoleToDisplayName map[string]string
	space                  models.Space
}

func init() {
	command_registry.Register(&SpaceUsers{})
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
	cmd.pluginModel = deps.PluginModels.SpaceUsers

	return cmd
}

func (cmd *SpaceUsers) Execute(c flags.FlagContext) {
	spaceName := c.Args()[1]
	org := cmd.orgReq.GetOrganization()

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	printer := cmd.getPrinter(space)
	printer.printUsers()
}

func (cmd *SpaceUsers) getPrinter(space models.Space) userPrinter {
	var spaceRoleToDisplayName = map[string]string{
		models.SPACE_MANAGER:   T("SPACE MANAGER"),
		models.SPACE_DEVELOPER: T("SPACE DEVELOPER"),
		models.SPACE_AUDITOR:   T("SPACE AUDITOR"),
	}

	if cmd.pluginCall {
		return &pluginPrinter{
			space:                  space,
			pluginModel:            cmd.pluginModel,
			usersMap:               make(map[string]plugin_models.GetSpaceUsers_Model),
			userLister:             cmd.getUserLister(),
			spaceRoleToDisplayName: spaceRoleToDisplayName,
		}
	}
	return &uiPrinter{
		ui:                     cmd.ui,
		space:                  space,
		userLister:             cmd.getUserLister(),
		spaceRoleToDisplayName: spaceRoleToDisplayName,
	}
}

func (cmd *SpaceUsers) getUserLister() func(spaceGuid string, role string) ([]models.UserFields, error) {
	if cmd.config.IsMinApiVersion("2.21.0") {
		return cmd.userRepo.ListUsersInSpaceForRoleWithNoUAA
	}
	return cmd.userRepo.ListUsersInSpaceForRole
}

func (p *pluginPrinter) printUsers() {
	for role, _ := range p.spaceRoleToDisplayName {
		users, _ := p.userLister(p.space.Guid, role)
		for _, user := range users {
			u, found := p.usersMap[user.Username]
			if !found {
				u = plugin_models.GetSpaceUsers_Model{}
				u.Username = user.Username
				u.Guid = user.Guid
				u.IsAdmin = user.IsAdmin
				u.Roles = make([]string, 1)
				u.Roles[0] = role
				p.usersMap[user.Username] = u
			} else {
				u.Roles = append(u.Roles, role)
				p.usersMap[user.Username] = u
			}
		}

	}
	for _, v := range p.usersMap {
		*(p.pluginModel) = append(*(p.pluginModel), v)
	}
}

func (p *uiPrinter) printUsers() {
	for role, displayName := range p.spaceRoleToDisplayName {
		users, err := p.userLister(p.space.Guid, role)
		if err != nil {
			p.ui.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                  err.Error(),
					"SpaceRoleToDisplayName": displayName,
				}))
			return
		}
		p.ui.Say("")
		p.ui.Say("%s", terminal.HeaderColor(displayName))

		if len(users) == 0 {
			p.ui.Say("none")
		} else {
			for _, user := range users {
				p.ui.Say("  %s", user.Username)
			}
		}
	}
}
