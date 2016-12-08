package application

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"sync"

	"sync/atomic"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/logs"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

const (
	DefaultStagingTimeout = 15 * time.Minute
	DefaultStartupTimeout = 5 * time.Minute
	DefaultPingerThrottle = 5 * time.Second
)

const LogMessageTypeStaging = "STG"

//go:generate counterfeiter . StagingWatcher

type StagingWatcher interface {
	WatchStaging(app models.Application, orgName string, spaceName string, startCommand func(app models.Application) (models.Application, error)) (updatedApp models.Application, err error)
}

//go:generate counterfeiter . Starter

type Starter interface {
	commandregistry.Command
	SetStartTimeoutInSeconds(timeout int)
	ApplicationStart(app models.Application, orgName string, spaceName string) (updatedApp models.Application, err error)
}

type Start struct {
	ui               terminal.UI
	config           coreconfig.Reader
	appDisplayer     Displayer
	appReq           requirements.ApplicationRequirement
	appRepo          applications.Repository
	logRepo          logs.Repository
	appInstancesRepo appinstances.Repository

	LogServerConnectionTimeout time.Duration
	StartupTimeout             time.Duration
	StagingTimeout             time.Duration
	PingerThrottle             time.Duration
}

func init() {
	commandregistry.Register(&Start{})
}

func (cmd *Start) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "start",
		ShortName:   "st",
		Description: T("Start an app"),
		Usage: []string{
			T("CF_NAME start APP_NAME"),
		},
	}
}

func (cmd *Start) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("start"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *Start) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()
	cmd.logRepo = deps.RepoLocator.GetLogsRepository()
	cmd.LogServerConnectionTimeout = 20 * time.Second
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

	appCommand := commandregistry.Commands.FindCommand("app")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appDisplayer = appCommand.(Displayer)

	return cmd
}

func (cmd *Start) Execute(c flags.FlagContext) error {
	_, err := cmd.ApplicationStart(cmd.appReq.GetApplication(), cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	return err
}

func (cmd *Start) ApplicationStart(app models.Application, orgName, spaceName string) (models.Application, error) {
	if app.State == "started" {
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already started")))
		return models.Application{}, nil
	}

	return cmd.WatchStaging(app, orgName, spaceName, func(app models.Application) (models.Application, error) {
		cmd.ui.Say(T("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(orgName),
				"SpaceName":   terminal.EntityNameColor(spaceName),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

		state := "started"
		return cmd.appRepo.Update(app.GUID, models.AppParams{State: &state})
	})
}

func (cmd *Start) WatchStaging(app models.Application, orgName, spaceName string, start func(app models.Application) (models.Application, error)) (models.Application, error) {
	stopChan := make(chan bool, 1)

	loggingStartedWait := new(sync.WaitGroup)
	loggingStartedWait.Add(1)

	loggingDoneWait := new(sync.WaitGroup)
	loggingDoneWait.Add(1)

	go cmd.TailStagingLogs(app, stopChan, loggingStartedWait, loggingDoneWait)

	loggingStartedWait.Wait()

	updatedApp, err := start(app)
	if err != nil {
		return models.Application{}, err
	}

	isStaged, err := cmd.waitForInstancesToStage(updatedApp)
	if err != nil {
		return models.Application{}, err
	}

	stopChan <- true

	loggingDoneWait.Wait()

	cmd.ui.Say("")

	if !isStaged {
		return models.Application{}, fmt.Errorf("%s failed to stage within %f minutes", app.Name, cmd.StagingTimeout.Minutes())
	}

	if app.InstanceCount > 0 {
		err = cmd.waitForOneRunningInstance(updatedApp)
		if err != nil {
			return models.Application{}, err
		}
		cmd.ui.Say(terminal.HeaderColor(T("\nApp started\n")))
		cmd.ui.Say("")
	} else {
		cmd.ui.Say(terminal.HeaderColor(T("\nApp state changed to started, but note that it has 0 instances.\n")))
		cmd.ui.Say("")
	}
	cmd.ui.Ok()

	//detectedstartcommand on first push is not present until starting completes
	startedApp, err := cmd.appRepo.GetApp(updatedApp.GUID)
	if err != nil {
		return models.Application{}, err
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

	err = cmd.appDisplayer.ShowApp(startedApp, orgName, spaceName)
	if err != nil {
		return models.Application{}, err
	}

	return updatedApp, nil
}

func (cmd *Start) SetStartTimeoutInSeconds(timeout int) {
	cmd.StartupTimeout = time.Duration(timeout) * time.Second
}

type ConnectionType int

const (
	NoConnection ConnectionType = iota
	ConnectionWasEstablished
	ConnectionWasClosed
	StoppedTrying
)

func (cmd *Start) TailStagingLogs(app models.Application, stopChan chan bool, startWait, doneWait *sync.WaitGroup) {
	var connectionStatus atomic.Value
	connectionStatus.Store(NoConnection)

	var once sync.Once
	startWaitDone := func() {
		startWait.Done()
	}

	onConnect := func() {
		if connectionStatus.Load() != StoppedTrying {
			connectionStatus.Store(ConnectionWasEstablished)
			once.Do(startWaitDone)
		}
	}

	timer := time.NewTimer(cmd.LogServerConnectionTimeout)

	c := make(chan logs.Loggable)
	e := make(chan error)

	defer doneWait.Done()

	go cmd.logRepo.TailLogsFor(app.GUID, onConnect, c, e)

	for {
		select {
		case <-timer.C:
			if connectionStatus.Load() == NoConnection {
				connectionStatus.Store(StoppedTrying)
				cmd.ui.Warn("timeout connecting to log server, no log will be shown")
				once.Do(startWaitDone)
				return
			}
		case msg, ok := <-c:
			if !ok {
				return
			} else if msg.GetSourceName() == LogMessageTypeStaging {
				cmd.ui.Say(msg.ToSimpleLog())
			}

		case err, ok := <-e:
			if ok {
				if connectionStatus.Load() != ConnectionWasClosed {
					cmd.ui.Warn(T("Warning: error tailing logs"))
					cmd.ui.Say("%s", err)
					once.Do(startWaitDone)
					return
				}
			}

		case <-stopChan:
			if connectionStatus.Load() == ConnectionWasEstablished {
				connectionStatus.Store(ConnectionWasClosed)
				cmd.logRepo.Close()
			} else {
				return
			}
		}
	}
}

func (cmd *Start) waitForInstancesToStage(app models.Application) (bool, error) {
	stagingStartTime := time.Now()

	var err error

	if cmd.StagingTimeout == 0 {
		app, err = cmd.appRepo.GetApp(app.GUID)
	} else {
		for app.PackageState != "STAGED" && app.PackageState != "FAILED" && time.Since(stagingStartTime) < cmd.StagingTimeout {
			app, err = cmd.appRepo.GetApp(app.GUID)
			if err != nil {
				break
			}

			time.Sleep(cmd.PingerThrottle)
		}
	}

	if err != nil {
		return false, err
	}

	if app.PackageState == "FAILED" {
		cmd.ui.Say("")
		if app.StagingFailedReason == "NoAppDetectedError" {
			return false, errors.New(T(`{{.Err}}
			
TIP: Buildpacks are detected when the "{{.PushCommand}}" is executed from within the directory that contains the app source code.

Use '{{.BuildpackCommand}}' to see a list of supported buildpacks.

Use '{{.Command}}' for more in depth log information.`,
				map[string]interface{}{
					"Err":              app.StagingFailedReason,
					"PushCommand":      terminal.CommandColor(fmt.Sprintf("%s push", cf.Name)),
					"BuildpackCommand": terminal.CommandColor(fmt.Sprintf("%s buildpacks", cf.Name)),
					"Command":          terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name, app.Name))}))
		}
		return false, errors.New(T("{{.Err}}\n\nTIP: use '{{.Command}}' for more information",
			map[string]interface{}{
				"Err":     app.StagingFailedReason,
				"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name, app.Name))}))
	}

	if time.Since(stagingStartTime) >= cmd.StagingTimeout {
		return false, nil
	}

	return true, nil
}

func (cmd *Start) waitForOneRunningInstance(app models.Application) error {
	timer := time.NewTimer(cmd.StartupTimeout)

	for {
		select {
		case <-timer.C:
			tipMsg := T("Start app timeout\n\nTIP: Application must be listening on the right port. Instead of hard coding the port, use the $PORT environment variable.") + "\n\n"
			tipMsg += T("Use '{{.Command}}' for more information", map[string]interface{}{"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name, app.Name))})

			return errors.New(tipMsg)

		default:
			count, err := cmd.fetchInstanceCount(app.GUID)
			if err != nil {
				cmd.ui.Warn("Could not fetch instance count: %s", err.Error())
				time.Sleep(cmd.PingerThrottle)
				continue
			}

			cmd.ui.Say(instancesDetails(count))

			if count.running > 0 {
				return nil
			}

			if count.flapping > 0 || count.crashed > 0 {
				return fmt.Errorf(T("Start unsuccessful\n\nTIP: use '{{.Command}}' for more information",
					map[string]interface{}{"Command": terminal.CommandColor(fmt.Sprintf("%s logs %s --recent", cf.Name, app.Name))}))
			}

			time.Sleep(cmd.PingerThrottle)
		}
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

func (cmd Start) fetchInstanceCount(appGUID string) (instanceCount, error) {
	count := instanceCount{
		startingDetails: make(map[string]struct{}),
	}

	instances, apiErr := cmd.appInstancesRepo.GetInstances(appGUID)
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
			for d := range count.startingDetails {
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
