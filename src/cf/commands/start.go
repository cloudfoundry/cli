package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
	"time"
)

type Start struct {
	ui        term.UI
	config    *configuration.Configuration
	appRepo   api.ApplicationRepository
	startTime time.Time
	appReq    requirements.ApplicationRequirement
}

type ApplicationStarter interface {
	ApplicationStart(cf.Application)
}

func NewStart(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (s *Start) {
	s = new(Start)
	s.ui = ui
	s.config = config
	s.appRepo = appRepo

	return
}

func (s *Start) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	s.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []Requirement{&s.appReq}
	return
}

func (s *Start) Run(c *cli.Context) {
	s.ApplicationStart(s.appReq.Application)
}

func (s *Start) ApplicationStart(app cf.Application) {
	if app.State == "started" {
		s.ui.Say(term.Magenta("Application " + app.Name + " is already started."))
		return
	}

	s.ui.Say("Starting %s...", term.Cyan(app.Name))

	err := s.appRepo.Start(s.config, app)
	if err != nil {
		s.ui.Failed("Error starting application.", err)
		return
	}

	s.ui.Ok()

	instances, errorCode, err := s.appRepo.GetInstances(s.config, app)

	for err != nil {
		if errorCode != 170002 {
			s.ui.Say("")
			s.ui.Failed("Error staging application", err)
			return
		}

		s.ui.Wait(1)
		instances, errorCode, err = s.appRepo.GetInstances(s.config, app)
		s.ui.LoadingIndication()
	}

	s.ui.Say("")

	s.startTime = time.Now()

	for s.displayInstancesStatus(app, instances) {
		s.ui.Wait(1)
		instances, _, _ = s.appRepo.GetInstances(s.config, app)
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
		if len(app.Urls) == 0 {
			s.ui.Say("Start successful!")
		} else {
			s.ui.Say("Start successful! App %s available at %s", app.Name, app.Urls[0])
		}
		return false
	} else {
		details := instancesDetails(runningCount, startingCount, downCount)
		s.ui.Say("%d of %d instances running (%s)", runningCount, totalCount, details)
	}

	if time.Since(s.startTime) > s.config.ApplicationStartTimeout*time.Second {
		s.ui.Failed("Start app timeout", nil)
		return false
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
