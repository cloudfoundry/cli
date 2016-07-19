package commands

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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

func (cmd *Logout) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *Logout) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	return cmd
}

func (cmd *Logout) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Logging out..."))
	cmd.config.ClearSession()
	cmd.ui.Ok()
	return nil
}
