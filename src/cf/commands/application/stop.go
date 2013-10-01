package application

import (
	"cf"
	"cf/api"
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
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewStop(ui terminal.UI, appRepo api.ApplicationRepository) (cmd *Stop) {
	cmd = new(Stop)
	cmd.ui = ui
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
		cmd.ui.Say(terminal.WarningColor("Application " + app.Name + " is already stopped."))
		return
	}

	cmd.ui.Say("Stopping %s...", terminal.EntityNameColor(app.Name))

	updatedApp, apiErr := cmd.appRepo.Stop(app)
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
	cmd.ApplicationStop(app)
}
