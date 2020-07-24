package application

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SSHEnabled struct {
	ui     terminal.UI
	config coreconfig.Reader
	appReq requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&SSHEnabled{})
}

func (cmd *SSHEnabled) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "ssh-enabled",
		Description: T("Reports whether SSH is enabled on an application container instance"),
		Usage: []string{
			T("CF_NAME ssh-enabled APP_NAME"),
		},
	}
}

func (cmd *SSHEnabled) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("ssh-enabled"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *SSHEnabled) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	return cmd
}

func (cmd *SSHEnabled) Execute(fc flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	if app.EnableSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is enabled for")+" '%s'", app.Name))
	} else {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is disabled for")+" '%s'", app.Name))
	}

	cmd.ui.Say("")
	return nil
}
