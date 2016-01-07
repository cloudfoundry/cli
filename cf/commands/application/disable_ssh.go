package application

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type DisableSSH struct {
	ui      terminal.UI
	config  core_config.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.ApplicationRepository
}

func init() {
	command_registry.Register(&DisableSSH{})
}

func (cmd *DisableSSH) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "disable-ssh",
		Description: T("disable ssh for the application"),
		Usage:       T("CF_NAME disable-ssh APP_NAME"),
	}
}

func (cmd *DisableSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + command_registry.Commands.CommandUsage("disable-ssh"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *DisableSSH) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *DisableSSH) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	if !app.EnableSsh {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already disabled")+" for '%s'", app.Name))
		return
	}

	cmd.ui.Say(fmt.Sprintf(T("Disabling ssh support for '%s'..."), app.Name))
	cmd.ui.Say("")

	enable := false
	updatedApp, err := cmd.appRepo.Update(app.Guid, models.AppParams{EnableSsh: &enable})
	if err != nil {
		cmd.ui.Failed(T("Error disabling ssh support for ") + app.Name + ": " + err.Error())
	}

	if !updatedApp.EnableSsh {
		cmd.ui.Ok()
	} else {
		cmd.ui.Failed(T("ssh support is not disabled for ") + app.Name)
	}
}
