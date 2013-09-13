package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type DeleteSpace struct {
	ui        term.UI
	spaceRepo api.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewDeleteSpace(ui term.UI, sR api.SpaceRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.spaceRepo = sR
	return
}

func (cmd *DeleteSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	var spaceName string

	if len(c.Args()) == 1 {
		spaceName = c.Args()[0]
	}

	if spaceName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-space")
		return
	}

	cmd.spaceReq = reqFactory.NewSpaceRequirement(spaceName)

	reqs = []requirements.Requirement{cmd.spaceReq}
	return
}

func (cmd *DeleteSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	force := c.Bool("f")

	if !force {
		response := strings.ToLower(cmd.ui.Ask(
			"Really delete space %s and everything associated with it?%s",
			term.EntityNameColor(space.Name),
			term.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	cmd.ui.Say("Deleting space %s...", term.EntityNameColor(space.Name))
	err := cmd.spaceRepo.Delete(space)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	return
}
