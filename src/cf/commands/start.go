package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
)

type Start struct {
	ui      term.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
}

func NewStart(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (s Start) {
	s.ui = ui
	s.config = config
	s.appRepo = appRepo

	return
}

func (s Start) Run(c *cli.Context) {
	appName := c.Args()[0]
	app, err := s.appRepo.FindByName(s.config, appName)
	if err != nil {
		s.ui.Failed(fmt.Sprintf("Error finding application %s", term.Cyan(appName)), err)
		return
	}

	if app.State == "started" {
		s.ui.Say(term.Magenta("Application " + appName + " is already started."))
		return
	}

	s.ui.Say("Starting %s...", term.Cyan(appName))

	err = s.appRepo.Start(s.config, app)
	if err != nil {
		s.ui.Failed("Error starting application.", err)
		return
	}
	s.ui.Ok()
}
