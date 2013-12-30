package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultStagingTimeout = 20 * time.Minute
	DefaultStartupTimeout = 5 * time.Minute
	DefaultPingerThrottle = 5 * time.Second
)

const LogMessageTypeStaging = "STG"

type Start struct {
	ui               terminal.UI
	config           *configuration.Configuration
	appDisplayer     ApplicationDisplayer
	appReq           requirements.ApplicationRequirement
	appRepo          api.ApplicationRepository
	appInstancesRepo api.AppInstancesRepository
	logRepo          api.LogsRepository

	StartupTimeout time.Duration
	StagingTimeout time.Duration
	PingerThrottle time.Duration
}

type ApplicationStarter interface {
	SetStartTimeoutSeconds(timeout int)
	ApplicationStart(app cf.Application) (updatedApp cf.Application, err error)
}

func NewStart(ui terminal.UI, config *configuration.Configuration, appDisplayer ApplicationDisplayer, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (cmd *Start) {
	cmd = new(Start)
	cmd.ui = ui
	cmd.config = config
	cmd.appDisplayer = appDisplayer
	cmd.appRepo = appRepo
	cmd.appInstancesRepo = appInstancesRepo
	cmd.logRepo = logRepo

	cmd.PingerThrottle = DefaultPingerThrottle

	if os.Getenv("CF_STAGING_TIMEOUT") != "" {
		duration, err := strconv.ParseInt(os.Getenv("CF_STAGING_TIMEOUT"), 10, 64)
		if err != nil {
			cmd.ui.Failed("invalid value for env var CF_STAGING_TIMEOUT\n%s", err)
		}
		cmd.StagingTimeout = time.Duration(duration) * time.Minute
	} else {
		cmd.StagingTimeout = DefaultStagingTimeout
	}

	if os.Getenv("CF_STARTUP_TIMEOUT") != "" {
		duration, err := strconv.ParseInt(os.Getenv("CF_STARTUP_TIMEOUT"), 10, 64)
		if err != nil {
			cmd.ui.Failed("invalid value for env var CF_STARTUP_TIMEOUT\n%s", err)
		}
		cmd.StartupTimeout = time.Duration(duration) * time.Minute
	} else {
		cmd.StartupTimeout = DefaultStartupTimeout
	}

	return
}

func (cmd *Start) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "start")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{cmd.appReq}
	return
}

func (cmd *Start) Run(c *cli.Context) {
	cmd.ApplicationStart(cmd.appReq.GetApplication())
}

func (cmd *Start) ApplicationStart(app cf.Application) (updatedApp cf.Application, err error) {
	if app.State == "started" {
		cmd.ui.Say(terminal.WarningColor("App " + app.Name + " is already started"))
		return
	}

	stopLoggingChan := make(chan bool, 1)
	defer close(stopLoggingChan)
	loggingStartedChan := make(chan bool)
	defer close(loggingStartedChan)

	go cmd.tailStagingLogs(app, loggingStartedChan, stopLoggingChan)

	<-loggingStartedChan

	cmd.ui.Say("Starting app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	params := cf.NewEmptyAppParams()
	params.Set("state", "STARTED")

	updatedApp, apiResponse := cmd.appRepo.Update(app.Guid, params)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	cmd.waitForInstancesToStage(updatedApp)
	stopLoggingChan <- true

	cmd.ui.Say("")

	cmd.waitForOneRunningInstance(app.Guid)
	cmd.ui.Say(terminal.HeaderColor("\nApp started\n"))

	cmd.appDisplayer.ShowApp(app)
	return
}

func (cmd *Start) SetStartTimeoutSeconds(timeout int) {
	cmd.StartupTimeout = time.Duration(timeout) * time.Second
}

func (cmd Start) tailStagingLogs(app cf.Application, startChan chan bool, stopChan chan bool) {
	logChan := make(chan *logmessage.Message, 1000)
	go func() {
		defer close(logChan)

		onConnect := func() {
			startChan <- true
		}

		err := cmd.logRepo.TailLogsFor(app.Guid, onConnect, logChan, stopChan, 1)
		if err != nil {
			cmd.ui.Warn("Warning: error tailing logs")
			cmd.ui.Say("%s", err)
			startChan <- true
		}
	}()

	cmd.displayLogMessages(logChan)
}

func (cmd Start) displayLogMessages(logChan chan *logmessage.Message) {
	for msg := range logChan {
		if msg.GetLogMessage().GetSourceName() != "STG" {
			continue
		}
		cmd.ui.Say(simpleLogMessageOutput(msg))
	}
}

func (cmd Start) waitForInstancesToStage(app cf.Application) {
	stagingStartTime := time.Now()
	_, apiResponse := cmd.appInstancesRepo.GetInstances(app.Guid)

	for apiResponse.IsNotSuccessful() && time.Since(stagingStartTime) < cmd.StagingTimeout {
		if apiResponse.ErrorCode != cf.APP_NOT_STAGED {
			cmd.ui.Say("")
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		cmd.ui.Wait(cmd.PingerThrottle)
		_, apiResponse = cmd.appInstancesRepo.GetInstances(app.Guid)
	}
	return
}

func (cmd Start) waitForOneRunningInstance(appGuid string) {
	var runningCount, startingCount, flappingCount, downCount int
	startupStartTime := time.Now()

	for runningCount == 0 {
		if time.Since(startupStartTime) > cmd.StartupTimeout {
			cmd.ui.Failed("Start app timeout")
			return
		}

		instances, apiResponse := cmd.appInstancesRepo.GetInstances(appGuid)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Wait(cmd.PingerThrottle)
			continue
		}

		totalCount := len(instances)
		runningCount, startingCount, flappingCount, downCount = 0, 0, 0, 0

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

		cmd.ui.Say(instancesDetails(startingCount, downCount, runningCount, flappingCount, totalCount))

		if flappingCount > 0 {
			cmd.ui.Failed("Start unsuccessful")
			return
		}
	}
}

func instancesDetails(startingCount, downCount, runningCount, flappingCount, totalCount int) string {
	details := []string{fmt.Sprintf("%d of %d instances running", runningCount, totalCount)}

	if startingCount > 0 {
		details = append(details, fmt.Sprintf("%d starting", startingCount))
	}

	if downCount > 0 {
		details = append(details, fmt.Sprintf("%d down", downCount))
	}

	if flappingCount > 0 {
		details = append(details, fmt.Sprintf("%d failing", flappingCount))
	}

	return strings.Join(details, ", ")
}
