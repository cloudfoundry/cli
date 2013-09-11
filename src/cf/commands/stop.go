package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Stop struct {
	ui      term.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewStop(ui term.UI, appRepo api.ApplicationRepository) (s *Stop) {
	s = new(Stop)
	s.ui = ui
	s.appRepo = appRepo

	return
}

func (s *Stop) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		s.ui.FailWithUsage(c, "stop")
		return
	}

	s.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{s.appReq}
	return
}

func (s *Stop) Run(c *cli.Context) {
	app := s.appReq.GetApplication()

	if app.State == "stopped" {
		s.ui.Say(term.Magenta("Application " + app.Name + " is already stopped."))
		return
	}

	s.ui.Say("Stopping %s...", term.Cyan(app.Name))

	err := s.appRepo.Stop(app)
	if err != nil {
		s.ui.Failed(err.Error())
		return
	}
	s.ui.Ok()
}
