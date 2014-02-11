package space

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameSpace struct {
	ui        terminal.UI
	config    configuration.ReadWriter
	spaceRepo api.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewRenameSpace(ui terminal.UI, config configuration.ReadWriter, spaceRepo api.SpaceRepository) (cmd *RenameSpace) {
	cmd = new(RenameSpace)
	cmd.ui = ui
	cmd.config = config
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
		reqFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *RenameSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	newName := c.Args()[1]
	cmd.ui.Say("Renaming space %s to %s in org %s as %s...",
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(newName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.spaceRepo.Rename(space.Guid, newName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if cmd.config.SpaceFields().Guid == space.Guid {
		space.Name = newName
		cmd.config.SetSpaceFields(space.SpaceFields)
	}

	cmd.ui.Ok()
}
