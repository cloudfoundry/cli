package application

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type SSHEnabled struct {
	ui     terminal.UI
	config core_config.Reader
	appReq requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&SSHEnabled{})
}

func (cmd *SSHEnabled) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "ssh-enabled",
		Description: T("reports whether SSH is enabled on an application container instance"),
		Usage:       T("CF_NAME ssh-enabled APP_NAME"),
	}
}

func (cmd *SSHEnabled) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + command_registry.Commands.CommandUsage("ssh-enabled"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *SSHEnabled) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	return cmd
}

func (cmd *SSHEnabled) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	if app.EnableSsh {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is enabled for")+" '%s'", app.Name))
	} else {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is disabled for")+" '%s'", app.Name))
	}

	cmd.ui.Say("")
}
