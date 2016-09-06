package application

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SetHealthCheck struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.Repository
}

func init() {
	commandregistry.Register(&SetHealthCheck{})
}

func (cmd *SetHealthCheck) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-health-check",
		Description: T("Set health_check_type flag to either 'port' or 'none'"),
		Usage: []string{
			T("CF_NAME set-health-check APP_NAME 'port'|'none'"),
		},
	}
}

func (cmd *SetHealthCheck) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME and HEALTH_CHECK_TYPE as arguments\n\n") + commandregistry.Commands.CommandUsage("set-health-check"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	if fc.Args()[1] != "port" && fc.Args()[1] != "none" {
		cmd.ui.Failed(T(`Incorrect Usage. HEALTH_CHECK_TYPE must be "port" or "none"\n\n`) + commandregistry.Commands.CommandUsage("set-health-check"))
		return nil, fmt.Errorf("Incorrect usage: invalid healthcheck type")
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *SetHealthCheck) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *SetHealthCheck) Execute(fc flags.FlagContext) error {
	healthCheckType := fc.Args()[1]

	app := cmd.appReq.GetApplication()

	if app.HealthCheckType == healthCheckType {
		cmd.ui.Say(fmt.Sprintf("%s "+T("health_check_type is already set")+" to '%s'", app.Name, app.HealthCheckType))
		return nil
	}

	cmd.ui.Say(fmt.Sprintf(T(""), app.Name, healthCheckType))
	cmd.ui.Say(T("Updating {{.AppName}} health_check_type to '{{.HealthCheckType}}'",
		map[string]interface{}{
			"AppName":         app.Name,
			"HealthCheckType": healthCheckType,
		},
	))
	cmd.ui.Say("")

	updatedApp, err := cmd.appRepo.Update(app.GUID, models.AppParams{HealthCheckType: &healthCheckType})
	if err != nil {
		return errors.New(T("Error updating health_check_type for ") + app.Name + ": " + err.Error())
	}

	if updatedApp.HealthCheckType == healthCheckType {
		cmd.ui.Ok()
	} else {
		return errors.New(T("health_check_type is not set to ") + healthCheckType + T(" for ") + app.Name)
	}
	return nil
}
