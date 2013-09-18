package commands

import (
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

	reqs = []requirements.Requirement{r.appReq}
	return
}

func (r *Restart) Run(c *cli.Context) {
	app := r.appReq.GetApplication()
	r.stopper.ApplicationStop(app)
	r.starter.ApplicationStart(app)
}
