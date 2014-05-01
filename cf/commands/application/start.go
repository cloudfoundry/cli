package application

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultStagingTimeout = 15 * time.Minute
	DefaultStartupTimeout = 5 * time.Minute
	DefaultPingerThrottle = 5 * time.Second
)

const LogMessageTypeStaging = "STG"

type Start struct {
	ui               terminal.UI
	config           configuration.Reader
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
	SetStartTimeoutInSeconds(timeout int)
	ApplicationStart(app models.Application) (updatedApp models.Application, err error)
}

func NewStart(ui terminal.UI, config configuration.Reader, appDisplayer ApplicationDisplayer, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (cmd *Start) {
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

func (command *Start) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "start",
		ShortName:   "st",
		Description: "Start an app",
		Usage:       "CF_NAME start APP",
	}
}

func (cmd *Start) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "start")
		return
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.appReq}
	return
}

func (cmd *Start) Run(c *cli.Context) {
	cmd.ApplicationStart(cmd.appReq.GetApplication())
}

func (cmd *Start) ApplicationStart(app models.Application) (updatedApp models.Application, err error) {
	if app.State == "started" {
		cmd.ui.Say(terminal.WarningColor("App " + app.Name + " is already started"))
		return
	}

	stopLoggingChan := make(chan bool, 1)
	loggingStartedChan := make(chan bool)

	go cmd.tailStagingLogs(app, loggingStartedChan, stopLoggingChan)

	<-loggingStartedChan // block until we have established connection to Loggregator

	cmd.ui.Say("Starting app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	state := "STARTED"
	updatedApp, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	cmd.waitForInstancesToStage(updatedApp)
	stopLoggingChan <- true

	cmd.ui.Say("")

	cmd.waitForOneRunningInstance(updatedApp)
	cmd.ui.Say(terminal.HeaderColor("\nApp started\n"))

	cmd.appDisplayer.ShowApp(updatedApp)
	return
}

func (cmd *Start) SetStartTimeoutInSeconds(timeout int) {
	cmd.StartupTimeout = time.Duration(timeout) * time.Second
}

func simpleLogMessageOutput(logMsg *logmessage.LogMessage) (msgText string) {
	msgText = string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err != nil {
		return
	}
	msgText = reg.ReplaceAllString(msgText, "")
	return
}

func (cmd Start) tailStagingLogs(app models.Application, startChan chan bool, stopChan chan bool) {
	onConnect := func() {
		startChan <- true
	}

	err := cmd.logRepo.TailLogsFor(app.Guid, 5*time.Second, onConnect, func(msg *logmessage.LogMessage) {
		select {
		case <-stopChan:
			cmd.logRepo.Close()
		default:
			if msg.GetSourceName() == LogMessageTypeStaging {
				cmd.ui.Say(simpleLogMessageOutput(msg))
			}
		}
	})

	if err != nil {
		cmd.ui.Warn("Warning: error tailing logs")
		cmd.ui.Say("%s", err)
		startChan <- true
	}
}

func (cmd Start) waitForInstancesToStage(app models.Application) {
	stagingStartTime := time.Now()
	_, err := cmd.appInstancesRepo.GetInstances(app.Guid)

	for err != nil && time.Since(stagingStartTime) < cmd.StagingTimeout {
		if httpError, ok := err.(errors.HttpError); ok && httpError.ErrorCode() != errors.APP_NOT_STAGED {
			cmd.ui.Say("")
			cmd.ui.Failed(fmt.Sprintf("%s\n\nTIP: use '%s' for more information",
				httpError.Error(),
				terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))))
			return
		}
		cmd.ui.Wait(cmd.PingerThrottle)
		_, err = cmd.appInstancesRepo.GetInstances(app.Guid)
	}
	return
}

func (cmd Start) waitForOneRunningInstance(app models.Application) {
	var runningCount, startingCount, flappingCount, downCount int
	startupStartTime := time.Now()

	for runningCount == 0 {
		if time.Since(startupStartTime) > cmd.StartupTimeout {
			cmd.ui.Failed(fmt.Sprintf("Start app timeout\n\nTIP: use '%s' for more information", terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))))
			return
		}

		instances, apiErr := cmd.appInstancesRepo.GetInstances(app.Guid)
		if apiErr != nil {
			cmd.ui.Wait(cmd.PingerThrottle)
			continue
		}

		totalCount := len(instances)
		runningCount, startingCount, flappingCount, downCount = 0, 0, 0, 0

		for _, inst := range instances {
			switch inst.State {
			case models.InstanceRunning:
				runningCount++
			case models.InstanceStarting:
				startingCount++
			case models.InstanceFlapping:
				flappingCount++
			case models.InstanceDown:
				downCount++
			}
		}

		cmd.ui.Say(instancesDetails(startingCount, downCount, runningCount, flappingCount, totalCount))

		if flappingCount > 0 {
			cmd.ui.Failed(fmt.Sprintf("Start unsuccessful\n\nTIP: use '%s' for more information", terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))))
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
