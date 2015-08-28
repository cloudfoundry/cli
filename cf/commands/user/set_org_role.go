package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
)

type SetOrgRole struct {
	ui       terminal.UI
	config   core_config.Reader
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

func (cmd *SetOrgRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments\n\n") + command_registry.Commands.CommandUsage("set-org-role"))
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(fc.Args()[0])
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[1])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}
	return
}

func (cmd *SetOrgRole) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
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

	apiErr := cmd.userRepo.SetOrgRole(user.Guid, org.Guid, role)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
