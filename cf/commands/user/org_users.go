package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin/models"
)

var orgRoles = []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR}

type OrgUsers struct {
	ui          terminal.UI
	config      core_config.Reader
	orgReq      requirements.OrganizationRequirement
	userRepo    api.UserRepository
	pluginModel *[]plugin_models.User
	pluginCall  bool
}

func init() {
	command_registry.Register(&OrgUsers{})
}

func NewOrgUsers(ui terminal.UI, config core_config.Reader, userRepo api.UserRepository) (cmd *OrgUsers) {
	cmd = &OrgUsers{}
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (cmd *OrgUsers) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["a"] = &cliFlags.BoolFlag{Name: "a", Usage: T("List all users in the org")}

	return command_registry.CommandMetadata{
		Name:        "org-users",
		Description: T("Show org users by role"),
		Usage:       T("CF_NAME org-users ORG"),
		Flags:       fs,
	}
}

func (cmd *OrgUsers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed("Incorrect Usage. Requires an argument\n\n" + command_registry.Commands.CommandUsage("org-users"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *OrgUsers) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Users
	return cmd
}

func (cmd *OrgUsers) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganization()
	all := c.Bool("a")

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	roles := orgRoles
	if all {
		roles = []string{models.ORG_USER}
	}

	var orgRoleToDisplayName = map[string]string{
		models.ORG_USER:        T("USERS"),
		models.ORG_MANAGER:     T("ORG MANAGER"),
		models.BILLING_MANAGER: T("BILLING MANAGER"),
		models.ORG_AUDITOR:     T("ORG AUDITOR"),
	}

	var usersMap = make(map[string]plugin_models.User)
	var users []models.UserFields
	var apiErr error

	for _, role := range roles {
		displayName := orgRoleToDisplayName[role]

		if cmd.config.IsMinApiVersion("2.21.0") {
			users, apiErr = cmd.userRepo.ListUsersInOrgForRoleWithNoUAA(org.Guid, role)
		} else {
			users, apiErr = cmd.userRepo.ListUsersInOrgForRole(org.Guid, role)
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
			cmd.ui.Failed(T("Failed fetching org-users for role {{.OrgRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                apiErr.Error(),
					"OrgRoleToDisplayName": displayName,
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
