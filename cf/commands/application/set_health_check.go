package application

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/applications"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SetHealthCheck struct {
	ui      terminal.UI
	config  core_config.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.ApplicationRepository
}

func init() {
	command_registry.Register(&SetHealthCheck{})
}

func (cmd *SetHealthCheck) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "set-health-check",
		Description: T("set health_check_type flag to either 'port' or 'none'"),
		Usage:       T("CF_NAME set-health-check APP_NAME 'port'|'none'"),
	}
}

func (cmd *SetHealthCheck) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME and HEALTH_CHECK_TYPE as arguments\n\n") + command_registry.Commands.CommandUsage("set-health-check"))
	}

	if fc.Args()[1] != "port" && fc.Args()[1] != "none" {
		cmd.ui.Failed(T(`Incorrect Usage. HEALTH_CHECK_TYPE must be "port" or "none"\n\n`) + command_registry.Commands.CommandUsage("set-health-check"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *SetHealthCheck) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *SetHealthCheck) Execute(fc flags.FlagContext) {
	healthCheckType := fc.Args()[1]

	app := cmd.appReq.GetApplication()

	if app.HealthCheckType == healthCheckType {
		cmd.ui.Say(fmt.Sprintf("%s "+T("health_check_type is already set")+" to '%s'", app.Name, app.HealthCheckType))
		return
	}

	cmd.ui.Say(fmt.Sprintf(T("Updating %s health_check_type to '%s'"), app.Name, healthCheckType))
	cmd.ui.Say("")

	updatedApp, err := cmd.appRepo.Update(app.Guid, models.AppParams{HealthCheckType: &healthCheckType})
	if err != nil {
		cmd.ui.Failed(T("Error updating health_check_type for ") + app.Name + ": " + err.Error())
	}

	if updatedApp.HealthCheckType == healthCheckType {
		cmd.ui.Ok()
	} else {
		cmd.ui.Failed(T("health_check_type is not set to ") + healthCheckType + T(" for ") + app.Name)
	}
}
