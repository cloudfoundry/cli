package space

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type DeleteSpace struct {
	ui         terminal.UI
	spaceRepo  api.SpaceRepository
	configRepo configuration.ConfigurationRepository
	spaceReq   requirements.SpaceRequirement
}

func NewDeleteSpace(ui terminal.UI, sR api.SpaceRepository, cR configuration.ConfigurationRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.spaceRepo = sR
	cmd.configRepo = cR
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

	cmd.ui.Warn("Deleting space %s...", spaceName)

	space, apiErr := cmd.spaceRepo.FindByName(spaceName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if !space.IsFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Space %s was already deleted.", spaceName)
		return
	}

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

	apiErr = cmd.spaceRepo.Delete(space)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
		return
	}

	if config.Space.Name == spaceName {
		config.Space = cf.Space{}
		cmd.configRepo.Save(config)
		cmd.ui.Say("TIP: No space targeted. Use '%s target -s' to target a space.", cf.Name)
	}

	return
}
