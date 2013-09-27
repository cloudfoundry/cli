package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type DeleteSpace struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewDeleteSpace(ui terminal.UI, sR api.SpaceRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.spaceRepo = sR
	return
}

func (cmd *DeleteSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-space")
		return
	}

	return
}

func (cmd *DeleteSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]

	force := c.Bool("f")
	if !force {
		response := strings.ToLower(cmd.ui.Ask(
			"Really delete space %s and everything associated with it?%s",
			terminal.EntityNameColor(spaceName),
			terminal.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	cmd.ui.Warn("Deleting space %s...", terminal.EntityNameColor(spaceName))

	space, found, apiErr := cmd.spaceRepo.FindByName(spaceName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if !found {
		cmd.ui.Ok()
		cmd.ui.Say("Space %s was already deleted.", terminal.EntityNameColor(spaceName))
		return
	}

	err := cmd.spaceRepo.Delete(space)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	return
}
