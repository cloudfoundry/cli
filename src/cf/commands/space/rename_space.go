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
	ui         terminal.UI
	spaceRepo  api.SpaceRepository
	spaceReq   requirements.SpaceRequirement
	configRepo configuration.ConfigurationRepository
}

func NewRenameSpace(ui terminal.UI, spaceRepo api.SpaceRepository, configRepo configuration.ConfigurationRepository) (cmd *RenameSpace) {
	cmd = new(RenameSpace)
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	cmd.configRepo = configRepo
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
	newName := c.Args()[1]
	cmd.ui.Say("Renaming space %s...", terminal.EntityNameColor(space.Name))

	apiStatus := cmd.spaceRepo.Rename(space, newName)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
		return
	}

	if config.Space.Guid == space.Guid {
		config.Space.Name = newName
		cmd.configRepo.Save(config)
	}

	cmd.ui.Ok()
}
