package space

import (
	"errors"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command *DeleteSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-space",
		Description: "Delete a space",
		Usage:       "CF_NAME delete-space SPACE [-f]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
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
