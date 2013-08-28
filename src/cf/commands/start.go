package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
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

	for s.displayInstancesStatus(app, instances) {
		instances, _ = s.appRepo.GetInstances(s.config, app)
	}
}

func (s Start) displayInstancesStatus(app cf.Application, instances []cf.ApplicationInstance) (notFinished bool) {
	totalCount := len(instances)
	runningCount, startingCount, flappingCount, downCount := 0, 0, 0, 0

	for _, inst := range instances {
		switch inst.State {
		case cf.InstanceRunning:
			runningCount++
		case cf.InstanceStarting:
			startingCount++
		case cf.InstanceFlapping:
			flappingCount++
		case cf.InstanceDown:
			downCount++
		}
	}

	if flappingCount > 0 {
		s.ui.Failed("Start unsuccessful", nil)
		return false
	}

	anyInstanceRunning := runningCount > 0

	if anyInstanceRunning {
		s.ui.Say("Start successful! App %s available at %s", app.Name, app.Urls[0])
		return false
	} else {
		details := instancesDetails(runningCount, startingCount, downCount)
		s.ui.Say("%d of %d instances running (%s)", runningCount, totalCount, details)
	}

	return totalCount > runningCount
}

func instancesDetails(runningCount int, startingCount int, downCount int) string {
	details := []string{}

	if startingCount > 0 {
		details = append(details, fmt.Sprintf("%d starting", startingCount))
	}

	if downCount > 0 {
		details = append(details, fmt.Sprintf("%d down", downCount))
	}

	return strings.Join(details, ", ")
}
