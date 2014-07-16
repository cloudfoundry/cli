package application

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type ApplicationStagingWatcher interface {
	ApplicationWatchStaging(app models.Application, startCommand func(app models.Application) (models.Application, error)) (updatedApp models.Application, err error)
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
			cmd.ui.Failed(T("invalid value for env var CF_STAGING_TIMEOUT\n{{.Err}}",
				map[string]interface{}{"Err": err}))
		}
		cmd.StagingTimeout = time.Duration(duration) * time.Minute
	} else {
		cmd.StagingTimeout = DefaultStagingTimeout
	}

	if os.Getenv("CF_STARTUP_TIMEOUT") != "" {
		duration, err := strconv.ParseInt(os.Getenv("CF_STARTUP_TIMEOUT"), 10, 64)
		if err != nil {
			cmd.ui.Failed(T("invalid value for env var CF_STARTUP_TIMEOUT\n{{.Err}}",
				map[string]interface{}{"Err": err}))
		}
		cmd.StartupTimeout = time.Duration(duration) * time.Minute
	} else {
		cmd.StartupTimeout = DefaultStartupTimeout
	}

	return
}

func (cmd *Start) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "start",
		ShortName:   "st",
		Description: T("Start an app"),
		Usage:       T("CF_NAME start APP"),
	}
}

func (cmd *Start) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		cmd.ui.FailWithUsage(c)
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
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already started")))
		return
	}

	return cmd.ApplicationWatchStaging(app, func(app models.Application) (models.Application, error) {
		cmd.ui.Say(T("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

		state := "STARTED"
		return cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	})
}

func (cmd *Start) ApplicationWatchStaging(app models.Application, start func(app models.Application) (models.Application, error)) (updatedApp models.Application, err error) {
	stopLoggingChan := make(chan bool, 1)
	loggingStartedChan := make(chan bool)

	go cmd.tailStagingLogs(app, loggingStartedChan, stopLoggingChan)

	<-loggingStartedChan // block until we have established connection to Loggregator

	updatedApp, apiErr := start(app)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	cmd.waitForInstancesToStage(updatedApp)
	stopLoggingChan <- true

	cmd.ui.Say("")

	cmd.waitForOneRunningInstance(updatedApp)
	cmd.ui.Say(terminal.HeaderColor(T("\nApp started\n")))

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

	err := cmd.logRepo.TailLogsFor(app.Guid, onConnect, func(msg *logmessage.LogMessage) {
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
		cmd.ui.Warn(T("Warning: error tailing logs"))
		cmd.ui.Say("%s", err)
		startChan <- true
	}
}

func isStagingError(err error) bool {
	httpError, ok := err.(errors.HttpError)
	return ok && httpError.ErrorCode() == errors.APP_NOT_STAGED
}

func (cmd Start) waitForInstancesToStage(app models.Application) {
	stagingStartTime := time.Now()
	_, err := cmd.appInstancesRepo.GetInstances(app.Guid)

	for isStagingError(err) && time.Since(stagingStartTime) < cmd.StagingTimeout {
		cmd.ui.Wait(cmd.PingerThrottle)
		_, err = cmd.appInstancesRepo.GetInstances(app.Guid)
	}

	if err != nil && !isStagingError(err) {
		cmd.ui.Say("")
		cmd.ui.Failed(T("{{.Err}}\n\nTIP: use '{{.Command}}' for more information",
			map[string]interface{}{
				"Err":     err.Error(),
				"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))}))
	}

	return
}

func (cmd Start) waitForOneRunningInstance(app models.Application) {
	startupStartTime := time.Now()

	for {
		if time.Since(startupStartTime) > cmd.StartupTimeout {
			cmd.ui.Failed(fmt.Sprintf(T("Start app timeout\n\nTIP: use '{{.Command}}' for more information",
				map[string]interface{}{
					"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))})))
			return
		}

		count, err := cmd.fetchInstanceCount(app.Guid)
		if err != nil {
			cmd.ui.Wait(cmd.PingerThrottle)
			continue
		}

		cmd.ui.Say(instancesDetails(count))

		if count.running > 0 {
			return
		}

		if count.flapping > 0 {
			cmd.ui.Failed(fmt.Sprintf(T("Start unsuccessful\n\nTIP: use '{{.Command}}' for more information",
				map[string]interface{}{"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))})))
			return
		}

		cmd.ui.Wait(cmd.PingerThrottle)
	}
}

type instanceCount struct {
	running  int
	starting int
	flapping int
	down     int
	total    int
}

func (cmd Start) fetchInstanceCount(appGuid string) (instanceCount, error) {
	count := instanceCount{}

	instances, apiErr := cmd.appInstancesRepo.GetInstances(appGuid)
	if apiErr != nil {
		return instanceCount{}, apiErr
	}

	count.total = len(instances)

	for _, inst := range instances {
		switch inst.State {
		case models.InstanceRunning:
			count.running++
		case models.InstanceStarting:
			count.starting++
		case models.InstanceFlapping:
			count.flapping++
		case models.InstanceDown:
			count.down++
		}
	}

	return count, nil
}

func instancesDetails(count instanceCount) string {
	details := []string{fmt.Sprintf(T("{{.RunningCount}} of {{.TotalCount}} instances running",
		map[string]interface{}{"RunningCount": count.running, "TotalCount": count.total}))}

	if count.starting > 0 {
		details = append(details, fmt.Sprintf(T("{{.StartingCount}} starting",
			map[string]interface{}{"StartingCount": count.starting})))
	}

	if count.down > 0 {
		details = append(details, fmt.Sprintf(T("{{.DownCount}} down",
			map[string]interface{}{"DownCount": count.down})))
	}

	if count.flapping > 0 {
		details = append(details, fmt.Sprintf(T("{{.FlappingCount}} failing",
			map[string]interface{}{"FlappingCount": count.flapping})))
	}

	return strings.Join(details, ", ")
}
