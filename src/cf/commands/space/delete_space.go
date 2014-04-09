package space

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteSpace struct {
	ui        terminal.UI
	config    configuration.ReadWriter
	spaceRepo api.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewDeleteSpace(ui terminal.UI, config configuration.ReadWriter, spaceRepo api.SpaceRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd *DeleteSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-space")
		return
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *DeleteSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete("space", spaceName) {
			return
		}
	}

	cmd.ui.Say("Deleting space %s in org %s as %s...",
		terminal.EntityNameColor(spaceName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	space := cmd.spaceReq.GetSpace()

	apiErr := cmd.spaceRepo.Delete(space.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if cmd.config.SpaceFields().Name == spaceName {
		cmd.config.SetSpaceFields(models.SpaceFields{})
		cmd.ui.Say("TIP: No space targeted, use '%s target -s' to target a space", cf.Name())
	}

	return
}
