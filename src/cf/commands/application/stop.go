package application

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
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

func (cmd *Stop) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "stop")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{cmd.appReq}
	return
}

func (cmd *Stop) ApplicationStop(app models.Application) (updatedApp models.Application, err error) {
	if app.State == "stopped" {
		updatedApp = app
		cmd.ui.Say(terminal.WarningColor("App " + app.Name + " is already stopped"))
		return
	}

	cmd.ui.Say("Stopping app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	state := "STOPPED"
	updatedApp, apiResponse := cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	if apiResponse != nil {
		err = errors.New(apiResponse.Error())
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd *Stop) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ApplicationStop(app)
}
