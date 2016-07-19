package user

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

//go:generate counterfeiter . OrgRoleSetter

type OrgRoleSetter interface {
	commandregistry.Command
	SetOrgRole(orgGUID string, role models.Role, userGUID, userName string) error
}

type SetOrgRole struct {
	ui       terminal.UI
	config   coreconfig.Reader
	flagRepo featureflags.FeatureFlagRepository
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	orgReq   requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&SetOrgRole{})
}

func (cmd *SetOrgRole) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-org-role",
		Description: T("Assign an org role to a user"),
		Usage: []string{
			T("CF_NAME set-org-role USERNAME ORG ROLE\n\n"),
			T("ROLES:\n"),
			fmt.Sprintf("   'OrgManager' - %s", T("Invite and manage users, select and change plans, and set spending limits\n")),
			fmt.Sprintf("   'BillingManager' - %s", T("Create and manage the billing account and payment info\n")),
			fmt.Sprintf("   'OrgAuditor' - %s", T("Read-only access to org info and reports\n")),
		},
	}
}

func (cmd *SetOrgRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments\n\n") + commandregistry.Commands.CommandUsage("set-org-role"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 3)
	}

	var wantGUID bool
	if cmd.config.IsMinAPIVersion(cf.SetRolesByUsernameMinimumAPIVersion) {
		setRolesByUsernameFlag, err := cmd.flagRepo.FindByName("set_roles_by_username")
		wantGUID = (err != nil || !setRolesByUsernameFlag.Enabled)
	} else {
		wantGUID = true
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(fc.Args()[0], wantGUID)
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[1])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *SetOrgRole) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *SetOrgRole) Execute(c flags.FlagContext) error {
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	roleStr := c.Args()[2]
	role, err := models.RoleFromString(roleStr)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Assigning role {{.Role}} to user {{.TargetUser}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(roleStr),
			"TargetUser":  terminal.EntityNameColor(user.Username),
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.SetOrgRole(org.GUID, role, user.GUID, user.Username)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}

func (cmd *SetOrgRole) SetOrgRole(orgGUID string, role models.Role, userGUID, userName string) error {
	if len(userGUID) > 0 {
		return cmd.userRepo.SetOrgRoleByGUID(userGUID, orgGUID, role)
	}

	return cmd.userRepo.SetOrgRoleByUsername(userName, orgGUID, role)
}
