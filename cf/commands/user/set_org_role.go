package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SetOrgRole struct {
	ui       terminal.UI
	config   configuration.Reader
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	orgReq   requirements.OrganizationRequirement
}

func NewSetOrgRole(ui terminal.UI, config configuration.Reader, userRepo api.UserRepository) (cmd *SetOrgRole) {
	cmd = new(SetOrgRole)
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (cmd *SetOrgRole) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-org-role",
		Description: T("Assign an org role to a user"),
		Usage: T("CF_NAME set-org-role USERNAME ORG ROLE\n\n") +
			T("ROLES:\n") +
			T("   OrgManager - Invite and manage users, select and change plans, and set spending limits\n") +
			T("   BillingManager - Create and manage the billing account and payment info\n") +
			T("   OrgAuditor - Read-only access to org info and reports\n"),
	}
}

func (cmd *SetOrgRole) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(c.Args()[0])
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[1])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}

	return
}

func (cmd *SetOrgRole) Run(c *cli.Context) {
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
