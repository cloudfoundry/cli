package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetSpaceRole struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	userRepo  api.UserRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func NewSetSpaceRole(ui terminal.UI, spaceRepo api.SpaceRepository, userRepo api.UserRepository) (cmd *SetSpaceRole) {
	cmd = new(SetSpaceRole)
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *SetSpaceRole) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-space-role")
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

func (cmd *SetSpaceRole) Run(c *cli.Context) {
	spaceName := c.Args()[2]
	role := c.Args()[3]

	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	space, apiResponse := cmd.spaceRepo.FindByNameInOrg(spaceName, org)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Say("Assigning %s role to %s in %s space in %s org...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(user.Username),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(org.Name),
	)

	apiResponse = cmd.userRepo.SetSpaceRole(user, space, role)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
