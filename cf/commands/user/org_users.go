package user

import (
	"github.com/cloudfoundry/cli/cf/actors/userprint"
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

type OrgUsers struct {
	ui          terminal.UI
	config      core_config.Reader
	orgReq      requirements.OrganizationRequirement
	userRepo    api.UserRepository
	pluginModel *[]plugin_models.GetOrgUsers_Model
	pluginCall  bool
}

func init() {
	command_registry.Register(&OrgUsers{})
}

func (cmd *OrgUsers) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["a"] = &cliFlags.BoolFlag{ShortName: "a", Usage: T("List all users in the org")}

	return command_registry.CommandMetadata{
		Name:        "org-users",
		Description: T("Show org users by role"),
		Usage:       T("CF_NAME org-users ORG"),
		Flags:       fs,
	}
}

func (cmd *OrgUsers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("org-users"))
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
	cmd.pluginModel = deps.PluginModels.OrgUsers
	return cmd
}

func (cmd *OrgUsers) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	printer := cmd.printer(c)
	printer.PrintUsers(org.Guid, cmd.config.Username())
}

func (cmd *OrgUsers) printer(c flags.FlagContext) userprint.UserPrinter {
	var roles []string
	if c.Bool("a") {
		roles = []string{models.ORG_USER}
	} else {
		roles = []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR}
	}

	if cmd.pluginCall {
		return userprint.NewOrgUsersPluginPrinter(
			cmd.pluginModel,
			cmd.userLister(),
			roles,
		)
	}
	return &userprint.OrgUsersUiPrinter{
		Ui:         cmd.ui,
		UserLister: cmd.userLister(),
		Roles:      roles,
		RoleDisplayNames: map[string]string{
			models.ORG_USER:        T("USERS"),
			models.ORG_MANAGER:     T("ORG MANAGER"),
			models.BILLING_MANAGER: T("BILLING MANAGER"),
			models.ORG_AUDITOR:     T("ORG AUDITOR"),
		},
	}
}

func (cmd *OrgUsers) userLister() func(orgGuid string, role string) ([]models.UserFields, error) {
	if cmd.config.IsMinApiVersion("2.21.0") {
		return cmd.userRepo.ListUsersInOrgForRoleWithNoUAA
	}
	return cmd.userRepo.ListUsersInOrgForRole
}
