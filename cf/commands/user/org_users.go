package user

import (
	"fmt"

	"code.cloudfoundry.org/cli/v7/cf/actors/userprint"
	"code.cloudfoundry.org/cli/v7/cf/api"
	"code.cloudfoundry.org/cli/v7/cf/commandregistry"
	"code.cloudfoundry.org/cli/v7/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v7/cf/flags"
	. "code.cloudfoundry.org/cli/v7/cf/i18n"
	"code.cloudfoundry.org/cli/v7/cf/models"
	"code.cloudfoundry.org/cli/v7/cf/requirements"
	"code.cloudfoundry.org/cli/v7/cf/terminal"
	"code.cloudfoundry.org/cli/v7/plugin/models"
)

type OrgUsers struct {
	ui          terminal.UI
	config      coreconfig.Reader
	orgReq      requirements.OrganizationRequirement
	userRepo    api.UserRepository
	pluginModel *[]plugin_models.GetOrgUsers_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&OrgUsers{})
}

func (cmd *OrgUsers) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["a"] = &flags.BoolFlag{ShortName: "a", Usage: T("List all users in the org")}

	return commandregistry.CommandMetadata{
		Name:        "org-users",
		Description: T("Show org users by role"),
		Usage: []string{
			T("CF_NAME org-users ORG"),
		},
		Flags: fs,
	}
}

func (cmd *OrgUsers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("org-users"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *OrgUsers) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.OrgUsers
	return cmd
}

func (cmd *OrgUsers) Execute(c flags.FlagContext) error {
	org := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	printer := cmd.printer(c)
	printer.PrintUsers(org.GUID, cmd.config.Username())
	return nil
}

func (cmd *OrgUsers) printer(c flags.FlagContext) userprint.UserPrinter {
	var roles []models.Role
	if c.Bool("a") {
		roles = []models.Role{models.RoleOrgUser}
	} else {
		roles = []models.Role{models.RoleOrgManager, models.RoleBillingManager, models.RoleOrgAuditor}
	}

	if cmd.pluginCall {
		return userprint.NewOrgUsersPluginPrinter(
			cmd.pluginModel,
			cmd.userRepo.ListUsersInOrgForRoleWithNoUAA,
			roles,
		)
	}
	return &userprint.OrgUsersUIPrinter{
		UI:         cmd.ui,
		UserLister: cmd.userRepo.ListUsersInOrgForRoleWithNoUAA,
		Roles:      roles,
		RoleDisplayNames: map[models.Role]string{
			models.RoleOrgUser:        T("USERS"),
			models.RoleOrgManager:     T("ORG MANAGER"),
			models.RoleBillingManager: T("BILLING MANAGER"),
			models.RoleOrgAuditor:     T("ORG AUDITOR"),
		},
	}
}
