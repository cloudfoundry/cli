package space

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteSpace struct {
	ui         terminal.UI
	config     *configuration.Configuration
	spaceRepo  api.SpaceRepository
	configRepo configuration.ConfigurationRepository
	spaceReq   requirements.SpaceRequirement
}

func NewDeleteSpace(ui terminal.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository, configRepo configuration.ConfigurationRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.configRepo = configRepo
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

	cmd.ui.Say("Deleting space %s in org %s as %s...",
		terminal.EntityNameColor(spaceName),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)

	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Space %s does not exist.", spaceName)
		return
	}

	if !force {
		response := cmd.ui.Confirm(
			"Really delete space %s and everything associated with it?%s",
			terminal.EntityNameColor(spaceName),
			terminal.PromptColor(">"),
		)
		if !response {
			return
		}
	}

	apiResponse = cmd.spaceRepo.Delete(space)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
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
		cmd.configRepo.Save()
		cmd.ui.Say("TIP: No space targeted, use '%s target -s' to target a space", cf.Name())
	}

	return
}
