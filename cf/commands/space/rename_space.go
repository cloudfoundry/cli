package space

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command *RenameSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-space",
		Description: "Rename a space",
		Usage:       "CF_NAME rename-space SPACE NEW_SPACE",
	}
}

func (cmd *RenameSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-space")
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

func (cmd *RenameSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	newName := c.Args()[1]
	cmd.ui.Say("Renaming space %s to %s in org %s as %s...",
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(newName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr := cmd.spaceRepo.Rename(space.Guid, newName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if cmd.config.SpaceFields().Guid == space.Guid {
		space.Name = newName
		cmd.config.SetSpaceFields(space.SpaceFields)
	}

	cmd.ui.Ok()
}
