package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ApplicationStopper interface {
	ApplicationStop(app cf.Application) (updatedApp cf.Application, err error)
}

type Stop struct {
	ui      terminal.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewStop(ui terminal.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (cmd *Stop) {
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

func (cmd *Stop) ApplicationStop(app cf.Application) (updatedApp cf.Application, err error) {
	if app.State == "stopped" {
		updatedApp = app
		cmd.ui.Say(terminal.WarningColor("App " + app.Name + " is already stopped"))
		return
	}

	cmd.ui.Say("Stopping %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	updatedApp, apiResponse := cmd.appRepo.Stop(app)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd *Stop) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ApplicationStop(app)
}
