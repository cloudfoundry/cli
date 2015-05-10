package application

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/noaa/events"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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
	config           core_config.Reader
	appDisplayer     ApplicationDisplayer
	appReq           requirements.ApplicationRequirement
	appRepo          applications.ApplicationRepository
	appInstancesRepo app_instances.AppInstancesRepository
	oldLogsRepo      api.OldLogsRepository
	logRepo          api.LogsNoaaRepository

	StartupTimeout time.Duration
	StagingTimeout time.Duration
	PingerThrottle time.Duration
}

type ApplicationStarter interface {
	SetStartTimeoutInSeconds(timeout int)
	ApplicationStart(app models.Application, orgName string, spaceName string) (updatedApp models.Application, err error)
}

type ApplicationStagingWatcher interface {
	ApplicationWatchStaging(app models.Application, orgName string, spaceName string, startCommand func(app models.Application) (models.Application, error)) (updatedApp models.Application, err error)
}

func NewStart(ui terminal.UI, config core_config.Reader, appDisplayer ApplicationDisplayer, appRepo applications.ApplicationRepository, appInstancesRepo app_instances.AppInstancesRepository, logRepo api.LogsNoaaRepository, oldLogsRepo api.OldLogsRepository) (cmd *Start) {
	cmd = new(Start)
	cmd.ui = ui
	cmd.config = config
	cmd.appDisplayer = appDisplayer
	cmd.appRepo = appRepo
	cmd.appInstancesRepo = appInstancesRepo
	cmd.logRepo = logRepo
	cmd.oldLogsRepo = oldLogsRepo

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
		Usage:       T("CF_NAME start APP_NAME"),
	}
}

func (cmd *Start) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(), cmd.appReq}
	return
}

func (cmd *Start) Run(c *cli.Context) {
	cmd.ApplicationStart(cmd.appReq.GetApplication(), cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
}

func (cmd *Start) ApplicationStart(app models.Application, orgName, spaceName string) (updatedApp models.Application, err error) {
	if app.State == "started" {
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already started")))
		return
	}

	return cmd.ApplicationWatchStaging(app, orgName, spaceName, func(app models.Application) (models.Application, error) {
		cmd.ui.Say(T("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(orgName),
				"SpaceName":   terminal.EntityNameColor(spaceName),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

		state := "STARTED"
		return cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	})
}

func (cmd *Start) ApplicationWatchStaging(app models.Application, orgName, spaceName string, start func(app models.Application) (models.Application, error)) (updatedApp models.Application, err error) {
	var isConnected bool
	loggingStartedChan := make(chan bool)
	doneLoggingChan := make(chan bool)

	go cmd.tailStagingLogs(app, loggingStartedChan, doneLoggingChan)
	timeout := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Second)
		close(timeout)
	}()

	select {
	case <-timeout:
		cmd.ui.Warn("timeout connecting to log server, no log will be shown")
		break
	case <-loggingStartedChan: // block until we have established connection to Loggregator
		isConnected = true
		break
	}

	updatedApp, apiErr := start(app)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	isStaged := cmd.waitForInstancesToStage(updatedApp)

	if isConnected { //only close when actually connected, else CLI hangs at closing consumer connection
		// cmd.logRepo.Close()
		cmd.oldLogsRepo.Close()
	}

	<-doneLoggingChan

	cmd.ui.Say("")

	if !isStaged {
		cmd.ui.Failed(fmt.Sprintf("%s failed to stage within %f minutes", app.Name, cmd.StagingTimeout.Minutes()))
	}

	cmd.waitForOneRunningInstance(updatedApp)
	cmd.ui.Say(terminal.HeaderColor(T("\nApp started\n")))
	cmd.ui.Say("")
	cmd.ui.Ok()

	//detectedstartcommand on first push is not present until starting completes
	startedApp, apiErr := cmd.appRepo.Read(updatedApp.Name)
	if err != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	var appStartCommand string
	if app.Command == "" {
		appStartCommand = startedApp.DetectedStartCommand
	} else {
		appStartCommand = startedApp.Command
	}

	cmd.ui.Say(T("\nApp {{.AppName}} was started using this command `{{.Command}}`\n",
		map[string]interface{}{
			"AppName": terminal.EntityNameColor(startedApp.Name),
			"Command": appStartCommand,
		}))

	cmd.appDisplayer.ShowApp(startedApp, orgName, spaceName)
	return
}

func (cmd *Start) SetStartTimeoutInSeconds(timeout int) {
	cmd.StartupTimeout = time.Duration(timeout) * time.Second
}

func simpleOldLogMessageOutput(logMsg *logmessage.LogMessage) (msgText string) {
	msgText = string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err != nil {
		return
	}
	msgText = reg.ReplaceAllString(msgText, "")
	return
}

func simpleLogMessageOutput(logMsg *events.LogMessage) (msgText string) {
	msgText = string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err != nil {
		return
	}
	msgText = reg.ReplaceAllString(msgText, "")
	return
}

func (cmd Start) tailStagingLogs(app models.Application, startChan, doneChan chan bool) {
	onConnect := func() {
		startChan <- true
	}

	err := cmd.oldLogsRepo.TailLogsFor(app.Guid, onConnect, func(msg *logmessage.LogMessage) {
		if msg.GetSourceName() == LogMessageTypeStaging {
			cmd.ui.Say(simpleOldLogMessageOutput(msg))
		}
	})
	// err := cmd.logRepo.TailNoaaLogsFor(app.Guid, onConnect, func(msg *events.LogMessage) {
	// 	if msg.GetSourceType() == LogMessageTypeStaging {
	// 		cmd.ui.Say(simpleLogMessageOutput(msg))
	// 	}
	// })

	if err != nil {
		cmd.ui.Warn(T("Warning: error tailing logs"))
		cmd.ui.Say("%s", err)
		startChan <- true
	}

	close(doneChan)
}

func (cmd Start) waitForInstancesToStage(app models.Application) bool {
	stagingStartTime := time.Now()

	var err error

	if cmd.StagingTimeout == 0 {
		app, err = cmd.appRepo.GetApp(app.Guid)
	} else {
		for app.PackageState != "STAGED" && app.PackageState != "FAILED" && time.Since(stagingStartTime) < cmd.StagingTimeout {
			app, err = cmd.appRepo.GetApp(app.Guid)
			if err != nil {
				break
			}
			cmd.ui.Wait(cmd.PingerThrottle)
		}
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if app.PackageState == "FAILED" {
		cmd.ui.Say("")
		cmd.ui.Failed(T("{{.Err}}\n\nTIP: use '{{.Command}}' for more information",
			map[string]interface{}{
				"Err":     app.StagingFailedReason,
				"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))}))
	}

	if time.Since(stagingStartTime) >= cmd.StagingTimeout {
		return false
	}

	return true
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

		if count.flapping > 0 || count.crashed > 0 {
			cmd.ui.Failed(fmt.Sprintf(T("Start unsuccessful\n\nTIP: use '{{.Command}}' for more information",
				map[string]interface{}{"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name(), app.Name))})))
			return
		}

		cmd.ui.Wait(cmd.PingerThrottle)
	}
}

type instanceCount struct {
	running         int
	starting        int
	startingDetails map[string]struct{}
	flapping        int
	down            int
	crashed         int
	total           int
}

func (cmd Start) fetchInstanceCount(appGuid string) (instanceCount, error) {
	count := instanceCount{
		startingDetails: make(map[string]struct{}),
	}

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
			if inst.Details != "" {
				count.startingDetails[inst.Details] = struct{}{}
			}
		case models.InstanceFlapping:
			count.flapping++
		case models.InstanceDown:
			count.down++
		case models.InstanceCrashed:
			count.crashed++
		}
	}

	return count, nil
}

func instancesDetails(count instanceCount) string {
	details := []string{fmt.Sprintf(T("{{.RunningCount}} of {{.TotalCount}} instances running",
		map[string]interface{}{"RunningCount": count.running, "TotalCount": count.total}))}

	if count.starting > 0 {
		if len(count.startingDetails) == 0 {
			details = append(details, fmt.Sprintf(T("{{.StartingCount}} starting",
				map[string]interface{}{"StartingCount": count.starting})))
		} else {
			info := []string{}
			for d, _ := range count.startingDetails {
				info = append(info, d)
			}
			sort.Strings(info)
			details = append(details, fmt.Sprintf(T("{{.StartingCount}} starting ({{.Details}})",
				map[string]interface{}{
					"StartingCount": count.starting,
					"Details":       strings.Join(info, ", "),
				})))
		}
	}

	if count.down > 0 {
		details = append(details, fmt.Sprintf(T("{{.DownCount}} down",
			map[string]interface{}{"DownCount": count.down})))
	}

	if count.flapping > 0 {
		details = append(details, fmt.Sprintf(T("{{.FlappingCount}} failing",
			map[string]interface{}{"FlappingCount": count.flapping})))
	}

	if count.crashed > 0 {
		details = append(details, fmt.Sprintf(T("{{.CrashedCount}} crashed",
			map[string]interface{}{"CrashedCount": count.crashed})))
	}

	return strings.Join(details, ", ")
}
