package space

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RenameSpace struct {
	ui        terminal.UI
	config    coreconfig.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func init() {
	commandregistry.Register(&RenameSpace{})
}

func (cmd *RenameSpace) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename-space",
		Description: T("Rename a space"),
		Usage: []string{
			T("CF_NAME rename-space SPACE NEW_SPACE"),
		},
	}
}

func (cmd *RenameSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME NEW_SPACE_NAME as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *RenameSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *RenameSpace) Execute(c flags.FlagContext) error {
	space := cmd.spaceReq.GetSpace()
	newName := c.Args()[1]
	cmd.ui.Say(T("Renaming space {{.OldSpaceName}} to {{.NewSpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OldSpaceName": terminal.EntityNameColor(space.Name),
			"NewSpaceName": terminal.EntityNameColor(newName),
			"OrgName":      terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser":  terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.spaceRepo.Rename(space.GUID, newName)
	if err != nil {
		return err
	}

	if cmd.config.SpaceFields().GUID == space.GUID {
		space.Name = newName
		cmd.config.SetSpaceFields(space.SpaceFields)
	}

	cmd.ui.Ok()
	return err
}
