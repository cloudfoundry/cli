package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnsetSpaceRole struct {
	ui       terminal.UI
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	spaceReq requirements.SpaceRequirement
}

func NewUnsetSpaceRole(ui terminal.UI, userRepo api.UserRepository) (cmd *UnsetSpaceRole) {
	cmd = new(UnsetSpaceRole)
	cmd.ui = ui
	cmd.userRepo = userRepo

	return
}

func (cmd *UnsetSpaceRole) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "unset-space-role")
		return
	}

	cmd.userReq = reqFactory.NewUserRequirement(c.Args()[0])
	cmd.spaceReq = reqFactory.NewSpaceRequirement(c.Args()[1])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.spaceReq,
	}

	return
}

func (cmd *UnsetSpaceRole) Run(c *cli.Context) {
	user := cmd.userReq.GetUser()
	space := cmd.spaceReq.GetSpace()
	role := c.Args()[2]

	cmd.ui.Say("Removing %s role from %s in %s space...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(user.Username),
		terminal.EntityNameColor(space.Name),
	)

	apiResponse := cmd.userRepo.UnsetSpaceRole(user, space, role)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
