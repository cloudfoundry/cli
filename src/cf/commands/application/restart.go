package application

import (
	"cf"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Restart struct {
	ui      terminal.UI
	starter ApplicationStarter
	stopper ApplicationStopper
	appReq  requirements.ApplicationRequirement
}

type ApplicationRestarter interface {
	ApplicationRestart(app cf.Application)
}

func NewRestart(ui terminal.UI, starter ApplicationStarter, stopper ApplicationStopper) (cmd *Restart) {
	cmd = new(Restart)
	cmd.ui = ui
	cmd.starter = starter
	cmd.stopper = stopper
	return
}

func (cmd *Restart) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "restart")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Restart) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	cmd.ApplicationRestart(app)
}

func (cmd *Restart) ApplicationRestart(app cf.Application) {
	stoppedApp, err := cmd.stopper.ApplicationStop(app)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	_, err = cmd.starter.ApplicationStart(stoppedApp)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}
