package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetOrgRole struct {
	ui       terminal.UI
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	orgReq   requirements.OrganizationRequirement
}

func NewSetOrgRole(ui terminal.UI, userRepo api.UserRepository) (cmd *SetOrgRole) {
	cmd = new(SetOrgRole)
	cmd.ui = ui
	cmd.userRepo = userRepo
	return
}

func (cmd *SetOrgRole) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-org-role")
		return
	}

	cmd.userReq = reqFactory.NewUserRequirement(c.Args()[0])
	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[1])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}

	return
}

func (cmd *SetOrgRole) Run(c *cli.Context) {
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	role := c.Args()[2]

	cmd.ui.Say("Assigning %s role to %s in %s org...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(user.Username),
		terminal.EntityNameColor(org.Name),
	)

	apiResponse := cmd.userRepo.SetOrgRole(user, org, role)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
