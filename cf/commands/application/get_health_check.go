package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type GetHealthCheck struct {
	ui      terminal.UI
	config  core_config.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.ApplicationRepository
}

func init() {
	command_registry.Register(&GetHealthCheck{})
}

func (cmd *GetHealthCheck) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "get-health-check",
		Description: T("get the health_check_type value of an app"),
		Usage:       T("CF_NAME get-health-check APP_NAME"),
	}
}

func (cmd *GetHealthCheck) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + command_registry.Commands.CommandUsage("get-health-check"))
	}

	cmd.ui.Say(T("Getting health_check_type value for ") + terminal.EntityNameColor(fc.Args()[0]))
	cmd.ui.Say("")

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *GetHealthCheck) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *GetHealthCheck) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("health_check_type is ") + terminal.HeaderColor(app.HealthCheckType))
}
