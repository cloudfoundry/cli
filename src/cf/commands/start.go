package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
	"time"
)

type Start struct {
	ui        terminal.UI
	config    *configuration.Configuration
	appRepo   api.ApplicationRepository
	startTime time.Time
	appReq    requirements.ApplicationRequirement
}

type ApplicationStarter interface {
	ApplicationStart(cf.Application) (startedApp cf.Application, err error)
}

func NewStart(ui terminal.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (s *Start) {
	s = new(Start)
	s.ui = ui
	s.config = config
	s.appRepo = appRepo

	return
}

func (s *Start) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		s.ui.FailWithUsage(c, "start")
		return
	}

	s.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{s.appReq}
	return
}

func (s *Start) Run(c *cli.Context) {
	s.ApplicationStart(s.appReq.GetApplication())
}

func (s *Start) ApplicationStart(app cf.Application) (updatedApp cf.Application, err error) {
	if app.State == "started" {
		s.ui.Say(terminal.WarningColor("Application " + app.Name + " is already started."))
		return
	}

	s.ui.Say("Starting %s...", terminal.EntityNameColor(app.Name))

	updatedApp, apiErr := s.appRepo.Start(app)
	if apiErr != nil {
		s.ui.Failed(apiErr.Error())
		return
	}

	s.ui.Ok()

	instances, apiErr := s.appRepo.GetInstances(app)

	for apiErr != nil {
		if apiErr.ErrorCode != net.APP_NOT_STAGED {
			s.ui.Say("")
			s.ui.Failed(apiErr.Error())
			return
		}

		s.ui.Wait(1 * time.Second)
		instances, apiErr = s.appRepo.GetInstances(app)
		s.ui.LoadingIndication()
	}

	s.ui.Say("")

	s.startTime = time.Now()

	for s.displayInstancesStatus(app, instances) {
		s.ui.Wait(1 * time.Second)
		instances, _ = s.appRepo.GetInstances(app)
	}

	return
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
		s.ui.Failed("Start unsuccessful")
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
		s.ui.Failed("Start app timeout")
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
