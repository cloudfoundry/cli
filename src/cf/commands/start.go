package commands

import (
	"cf"
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

	instances, err := s.appRepo.GetInstances(s.config, app)

	for err != nil {
		s.ui.Wait(1)
		instances, err = s.appRepo.GetInstances(s.config, app)
		s.ui.LoadingIndication()
	}

	s.ui.Say("")

	for s.displayInstancesStatus(instances) {
		instances, _ = s.appRepo.GetInstances(s.config, app)
	}
}

func (s Start) displayInstancesStatus(instances []cf.ApplicationInstance) (notFinished bool) {
	totalCount := len(instances)
	runningCount := 0
	startingCount := 0

	for _, inst := range instances {
		switch inst.State {
		case "running":
			runningCount++
		case "starting":
			startingCount++
		}
	}

	if runningCount < totalCount {
		s.ui.Say("%d of %d instances running (%d running, %d starting)", runningCount, totalCount, runningCount, startingCount)
	} else {
		s.ui.Say("%d of %d instances running", runningCount, totalCount)
	}

	return totalCount > runningCount
}
