package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type OrgUsers struct {
	ui       terminal.UI
	orgReq   requirements.OrganizationRequirement
	userRepo api.UserRepository
}

func NewOrgUsers(ui terminal.UI, userRepo api.UserRepository) (cmd *OrgUsers) {
	cmd = new(OrgUsers)
	cmd.ui = ui
	cmd.userRepo = userRepo
	return
}

func (cmd *OrgUsers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "org-users")
		return
	}

	orgName := c.Args()[0]
	cmd.orgReq = reqFactory.NewOrganizationRequirement(orgName)
	reqs = append(reqs, reqFactory.NewLoginRequirement(), cmd.orgReq)

	return
}

func (cmd *OrgUsers) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()

	cmd.ui.Say("Getting users in org %s...", terminal.EntityNameColor(org.Name))

	usersByRole, apiResponse := cmd.userRepo.FindAllInOrgByRole(org)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	for role, users := range usersByRole {
		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(role))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)
		}
	}
}
