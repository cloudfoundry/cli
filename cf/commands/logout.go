package commands

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Logout struct {
	ui     terminal.UI
	config coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(&Logout{})
}

func (cmd *Logout) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "logout",
		ShortName:   "lo",
		Description: T("Log user out"),
		Usage: []string{
			T("CF_NAME logout"),
		},
	}
}

func (cmd *Logout) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *Logout) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	return cmd
}

func (cmd *Logout) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Logging out..."))
	cmd.config.ClearSession()
	cmd.ui.Ok()
}
