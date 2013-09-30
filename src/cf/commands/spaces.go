package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Spaces struct {
	ui        terminal.UI
	config    configuration.Configuration
	spaceRepo api.SpaceRepository
}

func NewSpaces(ui terminal.UI, config configuration.Configuration, spaceRepo api.SpaceRepository) (cmd Spaces) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd Spaces) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd Spaces) Run(c *cli.Context) {
	cmd.ui.Say("Getting spaces in %s...", terminal.EntityNameColor(cmd.config.Organization.Name))

	spaces, err := cmd.spaceRepo.FindAll()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	for _, space := range spaces {
		cmd.ui.Say(space.Name)
	}
}
