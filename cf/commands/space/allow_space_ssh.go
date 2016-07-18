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

type AllowSpaceSSH struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceReq  requirements.SpaceRequirement
	spaceRepo spaces.SpaceRepository
}

func init() {
	commandregistry.Register(&AllowSpaceSSH{})
}

func (cmd *AllowSpaceSSH) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "allow-space-ssh",
		Description: T("Allow SSH access for the space"),
		Usage: []string{
			T("CF_NAME allow-space-ssh SPACE_NAME"),
		},
	}
}

func (cmd *AllowSpaceSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("allow-space-ssh"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *AllowSpaceSSH) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *AllowSpaceSSH) Execute(fc flags.FlagContext) error {
	space := cmd.spaceReq.GetSpace()

	if space.AllowSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already enabled in space ")+"'%s'", space.Name))
		return nil
	}

	cmd.ui.Say(fmt.Sprintf(T("Enabling ssh support for space '%s'..."), space.Name))
	cmd.ui.Say("")

	err := cmd.spaceRepo.SetAllowSSH(space.GUID, true)
	if err != nil {
		return errors.New(T("Error enabling ssh support for space ") + space.Name + ": " + err.Error())
	}

	cmd.ui.Ok()
	return nil
}
