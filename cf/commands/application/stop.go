package application

import (
	"errors"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ApplicationStopper interface {
	ApplicationStop(app models.Application) (updatedApp models.Application, err error)
}

type Stop struct {
	ui      terminal.UI
	config  configuration.Reader
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewStop(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository) (cmd *Stop) {
	cmd = new(Stop)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo

	return
}

func (cmd *Stop) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "stop",
		ShortName:   "sp",
		Description: T("Stop an app"),
		Usage:       T("CF_NAME stop APP"),
	}
}

func (cmd *Stop) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.appReq}
	return
}

func (cmd *Stop) ApplicationStop(app models.Application) (updatedApp models.Application, err error) {
	if app.State == "stopped" {
		updatedApp = app
		return
	}

	cmd.ui.Say(T("Stopping app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
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

func (cmd *Stop) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	if app.State == "stopped" {
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already stopped")))
	} else {
		cmd.ApplicationStop(app)
	}
}
