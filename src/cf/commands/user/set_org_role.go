package user

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
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
	role := models.UserInputToOrgRole[c.Args()[2]]

	cmd.ui.Say("Assigning role %s to user %s in org %s as %s...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(user.Username),
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.userRepo.SetOrgRole(user.Guid, org.Guid, role)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
