package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateSpace struct {
	ui        term.UI
	spaceRepo api.SpaceRepository
}

func NewCreateSpace(ui term.UI, sR api.SpaceRepository) (cmd CreateSpace) {
	cmd.ui = ui
	cmd.spaceRepo = sR
	return
}

func (cmd CreateSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-space")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd CreateSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	cmd.ui.Say("Creating space %s...", term.EntityNameColor(spaceName))

	err := cmd.spaceRepo.Create(spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\nTIP: Use '%s' to target new space.", term.CommandColor("cf target -s "+spaceName))
}
