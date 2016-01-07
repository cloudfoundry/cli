package user

import (
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

//go:generate counterfeiter -o fakes/fake_org_role_setter.go . OrgRoleSetter
type OrgRoleSetter interface {
	command_registry.Command
	SetOrgRole(orgGuid string, role, userGuid, userName string) error
}

type SetOrgRole struct {
	ui       terminal.UI
	config   core_config.Reader
	flagRepo feature_flags.FeatureFlagRepository
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	orgReq   requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&SetOrgRole{})
}

func (cmd *SetOrgRole) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "set-org-role",
		Description: T("Assign an org role to a user"),
		Usage: T("CF_NAME set-org-role USERNAME ORG ROLE\n\n") +
			T("ROLES:\n") +
			T("   OrgManager - Invite and manage users, select and change plans, and set spending limits\n") +
			T("   BillingManager - Create and manage the billing account and payment info\n") +
			T("   OrgAuditor - Read-only access to org info and reports\n"),
	}
}

func (cmd *SetOrgRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments\n\n") + command_registry.Commands.CommandUsage("set-org-role"))
	}

	var wantGuid bool
	if cmd.config.IsMinApiVersion("2.37.0") {
		setRolesByUsernameFlag, err := cmd.flagRepo.FindByName("set_roles_by_username")
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

	return reqs, nil
}

func (cmd *SetOrgRole) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *SetOrgRole) Execute(c flags.FlagContext) {
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	role := models.UserInputToOrgRole[c.Args()[2]]

	cmd.ui.Say(T("Assigning role {{.Role}} to user {{.TargetUser}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role),
			"TargetUser":  terminal.EntityNameColor(user.Username),
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.SetOrgRole(org.Guid, role, user.Guid, user.Username)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}

func (cmd *SetOrgRole) SetOrgRole(orgGuid string, role, userGuid, userName string) error {
	if len(userGuid) > 0 {
		return cmd.userRepo.SetOrgRoleByGuid(userGuid, orgGuid, role)
	}

	return cmd.userRepo.SetOrgRoleByUsername(userName, orgGuid, role)
}
