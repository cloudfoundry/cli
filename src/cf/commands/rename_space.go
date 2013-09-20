package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameSpace struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewRenameSpace(ui terminal.UI, spaceRepo api.SpaceRepository) (cmd *RenameSpace) {
	cmd = new(RenameSpace)
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd *RenameSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-space")
		return
	}
	cmd.spaceReq = reqFactory.NewSpaceRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *RenameSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	cmd.ui.Say("Renaming space %s...", terminal.EntityNameColor(space.Name))

	err := cmd.spaceRepo.Rename(space, c.Args()[1])
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()
}
