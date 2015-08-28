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

type UnsetOrgRole struct {
	ui       terminal.UI
	config   core_config.Reader
	userRepo api.UserRepository
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
		Usage: T("CF_NAME unset-org-role USERNAME ORG ROLE\n\n") +
			T("ROLES:\n") +
			T("   OrgManager - Invite and manage users, select and change plans, and set spending limits\n") +
			T("   BillingManager - Create and manage the billing account and payment info\n") +
			T("   OrgAuditor - Read-only access to org info and reports\n"),
	}
}

func (cmd *UnsetOrgRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments\n\n") + command_registry.Commands.CommandUsage("unset-org-role"))
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

func (cmd *UnsetOrgRole) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *UnsetOrgRole) Execute(c flags.FlagContext) {
	role := models.UserInputToOrgRole[c.Args()[2]]
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Removing role {{.Role}} from user {{.TargetUser}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role),
			"TargetUser":  terminal.EntityNameColor(c.Args()[0]),
			"TargetOrg":   terminal.EntityNameColor(c.Args()[1]),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	apiErr := cmd.userRepo.UnsetOrgRole(user.Guid, org.Guid, role)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
