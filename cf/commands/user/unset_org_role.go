package user

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/feature_flags"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type UnsetOrgRole struct {
	ui       terminal.UI
	config   core_config.Reader
	userRepo api.UserRepository
	flagRepo feature_flags.FeatureFlagRepository
	userReq  requirements.UserRequirement
	orgReq   requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&UnsetOrgRole{})
}

func (cmd *UnsetOrgRole) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "unset-org-role",
		Description: T("Remove an org role from a user"),
		Usage: []string{
			T("CF_NAME unset-org-role USERNAME ORG ROLE\n\n"),
			T("ROLES:\n"),
			fmt.Sprintf("   'OrgManager' - %s", T("Invite and manage users, select and change plans, and set spending limits\n")),
			fmt.Sprintf("   'BillingManager' - %s", T("Create and manage the billing account and payment info\n")),
			fmt.Sprintf("   'OrgAuditor' - %s", T("Read-only access to org info and reports\n")),
		},
	}
}

func (cmd *UnsetOrgRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments\n\n") + command_registry.Commands.CommandUsage("unset-org-role"))
	}

	var wantGuid bool
	if cmd.config.IsMinApiVersion(cf.SetRolesByUsernameMinimumApiVersion) {
		setRolesByUsernameFlag, err := cmd.flagRepo.FindByName("unset_roles_by_username")
		wantGuid = (err != nil || !setRolesByUsernameFlag.Enabled)
	} else {
		wantGuid = true
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(fc.Args()[0], wantGuid)
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[1])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}

	return reqs
}

func (cmd *UnsetOrgRole) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *UnsetOrgRole) Execute(c flags.FlagContext) {
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	role := models.UserInputToOrgRole[c.Args()[2]]

	cmd.ui.Say(T("Removing role {{.Role}} from user {{.TargetUser}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role),
			"TargetUser":  terminal.EntityNameColor(c.Args()[0]),
			"TargetOrg":   terminal.EntityNameColor(c.Args()[1]),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	var err error
	if len(user.Guid) > 0 {
		err = cmd.userRepo.UnsetOrgRoleByGuid(user.Guid, org.Guid, role)
	} else {
		err = cmd.userRepo.UnsetOrgRoleByUsername(user.Username, org.Guid, role)
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
