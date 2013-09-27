package commands

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

func (r *Restart) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		r.ui.FailWithUsage(c, "restart")
		return
	}

	r.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		r.appReq,
	}
	return
}

func (r *Restart) Run(c *cli.Context) {
	app := r.appReq.GetApplication()
	r.ApplicationRestart(app)
}

func (r *Restart) ApplicationRestart(app cf.Application) {
	stoppedApp, err := r.stopper.ApplicationStop(app)
	if err != nil {
		r.ui.Failed(err.Error())
		return
	}

	_, err = r.starter.ApplicationStart(stoppedApp)
	if err != nil {
		r.ui.Failed(err.Error())
		return
	}
}
