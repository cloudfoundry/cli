package space

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameSpace struct {
	ui        terminal.UI
	config    configuration.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewRenameSpace(ui terminal.UI, config configuration.ReadWriter, spaceRepo spaces.SpaceRepository) (cmd *RenameSpace) {
	cmd = new(RenameSpace)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd *RenameSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-space",
		Description: T("Rename a space"),
		Usage:       T("CF_NAME rename-space SPACE NEW_SPACE"),
	}
}

func (cmd *RenameSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
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
	cmd.ui.Say(T("Renaming space {{.OldSpaceName}} to {{.NewSpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OldSpaceName": terminal.EntityNameColor(space.Name),
			"NewSpaceName": terminal.EntityNameColor(newName),
			"OrgName":      terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser":  terminal.EntityNameColor(cmd.config.Username()),
		}))

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
