package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetSpaceRole struct {
	ui       terminal.UI
	userRepo api.UserRepository
	userReq  requirements.UserRequirement
	spaceReq requirements.SpaceRequirement
}

func NewSetSpaceRole(ui terminal.UI, userRepo api.UserRepository) (cmd *SetSpaceRole) {
	cmd = new(SetSpaceRole)
	cmd.ui = ui
	cmd.userRepo = userRepo

	return
}

func (cmd *SetSpaceRole) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-space-role")
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

func (cmd *SetSpaceRole) Run(c *cli.Context) {
	user := cmd.userReq.GetUser()
	space := cmd.spaceReq.GetSpace()
	role := c.Args()[2]

	cmd.ui.Say("Assigning %s role to %s in %s space...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(user.Username),
		terminal.EntityNameColor(space.Name),
	)

	apiResponse := cmd.userRepo.SetSpaceRole(user, space, role)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
