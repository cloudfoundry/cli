package space

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RenameSpace struct {
	ui        terminal.UI
	config    core_config.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func init() {
	command_registry.Register(&RenameSpace{})
}

func (cmd *RenameSpace) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "rename-space",
		Description: T("Rename a space"),
		Usage:       T("CF_NAME rename-space SPACE NEW_SPACE"),
	}
}

func (cmd *RenameSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME NEW_SPACE_NAME as arguments\n\n") + command_registry.Commands.CommandUsage("rename-space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *RenameSpace) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *RenameSpace) Execute(c flags.FlagContext) {
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
