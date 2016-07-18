package space

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SSHAllowed struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceReq  requirements.SpaceRequirement
	spaceRepo spaces.SpaceRepository
}

func init() {
	commandregistry.Register(&SSHAllowed{})
}

func (cmd *SSHAllowed) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "space-ssh-allowed",
		Description: T("Reports whether SSH is allowed in a space"),
		Usage: []string{
			T("CF_NAME space-ssh-allowed SPACE_NAME"),
		},
	}
}

func (cmd *SSHAllowed) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("space-ssh-allowed"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *SSHAllowed) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	return cmd
}

func (cmd *SSHAllowed) Execute(fc flags.FlagContext) error {
	space := cmd.spaceReq.GetSpace()

	if space.AllowSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is enabled in space ")+"'%s'", space.Name))
	} else {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is disabled in space ")+"'%s'", space.Name))
	}
	return nil
}
