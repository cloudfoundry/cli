package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
)

type Stop struct {
	ui      term.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
}

func NewStop(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (s Stop) {
	s.ui = ui
	s.config = config
	s.appRepo = appRepo

	return
}

func (s Stop) Run(c *cli.Context) {
	appName := c.Args()[0]
	app, err := s.appRepo.FindByName(s.config, appName)
	if err != nil {
		s.ui.Failed(fmt.Sprintf("Error finding application %s", term.Cyan(appName)), err)
		return
	}

	if app.State == "stopped" {
		s.ui.Say(term.Magenta("Application " + appName + " is already stopped."))
		return
	}

	s.ui.Say("Stopping %s...", term.Cyan(appName))

	err = s.appRepo.Stop(s.config, app)
	if err != nil {
		s.ui.Failed("Error stopping application.", err)
		return
	}
	s.ui.Ok()
}
