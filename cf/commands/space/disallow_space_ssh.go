package space

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DisallowSpaceSSH struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceReq  requirements.SpaceRequirement
	spaceRepo spaces.SpaceRepository
}

func init() {
	commandregistry.Register(&DisallowSpaceSSH{})
}

func (cmd *DisallowSpaceSSH) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "disallow-space-ssh",
		Description: T("Disallow SSH access for the space"),
		Usage: []string{
			T("CF_NAME disallow-space-ssh SPACE_NAME"),
		},
	}
}

func (cmd *DisallowSpaceSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("disallow-space-ssh"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *DisallowSpaceSSH) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *DisallowSpaceSSH) Execute(fc flags.FlagContext) error {
	space := cmd.spaceReq.GetSpace()

	if !space.AllowSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already disabled in space ")+"'%s'", space.Name))
		return nil
	}

	cmd.ui.Say(fmt.Sprintf(T("Disabling ssh support for space '%s'..."), space.Name))
	cmd.ui.Say("")

	err := cmd.spaceRepo.SetAllowSSH(space.GUID, false)
	if err != nil {
		return errors.New(T("Error disabling ssh support for space ") + space.Name + ": " + err.Error())
	}

	cmd.ui.Ok()
	return nil
}
