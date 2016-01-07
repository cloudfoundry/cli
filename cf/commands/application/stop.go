package application

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_application_stopper.go . ApplicationStopper
type ApplicationStopper interface {
	command_registry.Command
	ApplicationStop(app models.Application, orgName string, spaceName string) (updatedApp models.Application, err error)
}

type Stop struct {
	ui      terminal.UI
	config  core_config.Reader
	appRepo applications.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&Stop{})
}

func (cmd *Stop) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "stop",
		ShortName:   "sp",
		Description: T("Stop an app"),
		Usage:       T("CF_NAME stop APP_NAME"),
	}
}

func (cmd *Stop) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("stop"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(), cmd.appReq}

	return
}

func (cmd *Stop) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *Stop) ApplicationStop(app models.Application, orgName, spaceName string) (updatedApp models.Application, err error) {
	if app.State == "stopped" {
		updatedApp = app
		return
	}

	cmd.ui.Say(T("Stopping app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(orgName),
			"SpaceName":   terminal.EntityNameColor(spaceName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

	state := "STOPPED"
	updatedApp, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd *Stop) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()
	if app.State == "stopped" {
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already stopped")))
	} else {
		cmd.ApplicationStop(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}
}
