package space

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type AllowSpaceSSH struct {
	ui        terminal.UI
	config    core_config.Reader
	spaceReq  requirements.SpaceRequirement
	spaceRepo spaces.SpaceRepository
}

func init() {
	command_registry.Register(&AllowSpaceSSH{})
}

func (cmd *AllowSpaceSSH) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "allow-space-ssh",
		Description: T("allow SSH access for the space"),
		Usage:       T("CF_NAME allow-space-ssh SPACE_NAME"),
	}
}

func (cmd *AllowSpaceSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME as argument\n\n") + command_registry.Commands.CommandUsage("allow-space-ssh"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return
}

func (cmd *AllowSpaceSSH) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *AllowSpaceSSH) Execute(fc flags.FlagContext) {
	space := cmd.spaceReq.GetSpace()

	if space.AllowSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already enabled in space ")+"'%s'", space.Name))
		return
	}

	cmd.ui.Say(fmt.Sprintf(T("Enabling ssh support for space '%s'..."), space.Name))
	cmd.ui.Say("")

	err := cmd.spaceRepo.SetAllowSSH(space.Guid, true)
	if err != nil {
		cmd.ui.Failed(T("Error enabling ssh support for space ") + space.Name + ": " + err.Error())
	}

	cmd.ui.Ok()
}
