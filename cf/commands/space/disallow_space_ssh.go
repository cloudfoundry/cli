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

type DisallowSpaceSSH struct {
	ui        terminal.UI
	config    core_config.Reader
	spaceReq  requirements.SpaceRequirement
	spaceRepo spaces.SpaceRepository
}

func init() {
	command_registry.Register(&DisallowSpaceSSH{})
}

func (cmd *DisallowSpaceSSH) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "disallow-space-ssh",
		Description: T("disallow SSH access for the space"),
		Usage:       T("CF_NAME disallow-space-ssh SPACE_NAME"),
	}
}

func (cmd *DisallowSpaceSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME as argument\n\n") + command_registry.Commands.CommandUsage("disallow-space-ssh"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return
}

func (cmd *DisallowSpaceSSH) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *DisallowSpaceSSH) Execute(fc flags.FlagContext) {
	space := cmd.spaceReq.GetSpace()

	if !space.AllowSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already disabled in space ")+"'%s'", space.Name))
		return
	}

	cmd.ui.Say(fmt.Sprintf(T("Disabling ssh support for space '%s'..."), space.Name))
	cmd.ui.Say("")

	err := cmd.spaceRepo.SetAllowSSH(space.Guid, false)
	if err != nil {
		cmd.ui.Failed(T("Error disabling ssh support for space ") + space.Name + ": " + err.Error())
	}

	cmd.ui.Ok()
}
