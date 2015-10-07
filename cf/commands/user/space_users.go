package user

import (
	"github.com/cloudfoundry/cli/cf/actors/user_printer"
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

	printer := cmd.getPrinter()
	printer.PrintUsers(org, space, cmd.config.Username())
}

func (cmd *SpaceUsers) getPrinter() user_printer.UserPrinter {
	var roles = []string{models.SPACE_MANAGER, models.SPACE_DEVELOPER, models.SPACE_AUDITOR}

	if cmd.pluginCall {
		return &user_printer.PluginPrinter{
			PluginModel: cmd.pluginModel,
			UsersMap:    make(map[string]plugin_models.GetSpaceUsers_Model),
			UserLister:  cmd.getUserLister(),
			Roles:       roles,
		}
	}
	return &user_printer.UiPrinter{
		Ui:         cmd.ui,
		UserLister: cmd.getUserLister(),
		Roles:      roles,
		RoleDisplayNames: map[string]string{
			models.SPACE_MANAGER:   T("SPACE MANAGER"),
			models.SPACE_DEVELOPER: T("SPACE DEVELOPER"),
			models.SPACE_AUDITOR:   T("SPACE AUDITOR"),
		},
	}
}

func (cmd *SpaceUsers) getUserLister() func(spaceGuid string, role string) ([]models.UserFields, error) {
	if cmd.config.IsMinApiVersion("2.21.0") {
		return cmd.userRepo.ListUsersInSpaceForRoleWithNoUAA
	}
	return cmd.userRepo.ListUsersInSpaceForRole
}
