package v7

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v6shared "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/progressbar"

	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	v7pushaction.ProgressBar
	Complete()
	Ready()
}

//go:generate counterfeiter . PushActor

type PushActor interface {
	// Actualize applies any necessary changes.
	Actualize(state v7pushaction.PushState, progressBar v7pushaction.ProgressBar) (<-chan v7pushaction.PushState, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error)
	// Conceptualize figures out the state of the world.
	Conceptualize(appName string, spaceGUID string, orgGUID string, currentDir string, flagOverrides v7pushaction.FlagOverrides) ([]v7pushaction.PushState, v7pushaction.Warnings, error)
}

//go:generate counterfeiter . V7ActorForPush

type V7ActorForPush interface {
	AppActor
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	PollStart(appGUID string, warningsChannel chan<- v7action.Warnings) error
	RestartApplication(appGUID string) (v7action.Warnings, error)
}

type PushCommand struct {
	RequiredArgs        flag.AppName                `positional-args:"yes"`
	Buildpacks          []string                    `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	DockerImage         flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername      string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	HealthCheckType     flag.HealthCheckType        `long:"health-check-type" short:"u" description:"Application health check type: 'port' (default), 'process', 'http' (implies endpoint '/')"`
	Memory              flag.Megabytes              `long:"memory" short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoRoute             bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart             bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath             flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	StartCommand        flag.Command                `long:"start-command" short:"c" description:"Startup command, set to null to reset to default start command"`
	dockerPassword      interface{}                 `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage               interface{}                 `usage:"CF_NAME push APP_NAME [-b BUILDPACK]... [-p APP_PATH] [--no-route] [--no-start]\n   CF_NAME push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME] [--no-route] [--no-start]"`
	envCFStagingTimeout interface{}                 `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}                 `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI           command.UI
	Config       command.Config
	NOAAClient   v3action.NOAAClient
	Actor        PushActor
	VersionActor V7ActorForPush
	SharedActor  command.SharedActor
	RouteActor   v7action.RouteActor
	ProgressBar  ProgressBar
	PWD          string
}

func (cmd *PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.ProgressBar = progressbar.NewProgressBar()

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := v6shared.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		return err
	}

	v7actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.VersionActor = v7actor
	ccClientV2, uaaClientV2, err := v6shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)
	cmd.RouteActor = v2Actor
	cmd.Actor = v7pushaction.NewActor(v2Actor, v7actor, sharedActor)

	cmd.NOAAClient = v6shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

	currentDir, err := os.Getwd()
	cmd.PWD = currentDir

	return nil
}

func (cmd PushCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	overrides, err := cmd.GetFlagOverrides()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Pushing app {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	cmd.UI.DisplayText("Getting app info...")

	log.Info("generating the app state")
	pushState, warnings, err := cmd.Actor.Conceptualize(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		cmd.PWD,
		overrides,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	log.WithField("number of states", len(pushState)).Debug("completed generating state")

	for _, state := range pushState {
		log.WithField("app_name", state.Application.Name).Info("actualizing")
		stateStream, eventStream, warningsStream, errorStream := cmd.Actor.Actualize(state, cmd.ProgressBar)
		updatedState, err := cmd.processApplyStreams(state.Application.Name, stateStream, eventStream, warningsStream, errorStream)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for app to start...")
		warnings, err := cmd.VersionActor.RestartApplication(updatedState.Application.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		pollWarnings := make(chan v7action.Warnings)
		done := make(chan bool)
		go func() {
			for {
				select {
				case message := <-pollWarnings:
					cmd.UI.DisplayWarnings(message)
				case <-done:
					return
				}
			}
		}()

		err = cmd.VersionActor.PollStart(updatedState.Application.GUID, pollWarnings)
		log.Debug("blocking on 'done'")
		done <- true

		if err != nil {
			if _, ok := err.(actionerror.StartupTimeoutError); ok {
				return translatableerror.StartupTimeoutError{
					AppName:    cmd.RequiredArgs.AppName,
					BinaryName: cmd.Config.BinaryName(),
				}
			}

			return err
		}

		log.Info("getting application summary info")
		summary, warnings, err := cmd.VersionActor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, true, cmd.RouteActor)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
		appSummaryDisplayer.AppDisplay(summary, true)
	}

	return nil
}

func (cmd PushCommand) processApplyStreams(
	appName string,
	stateStream <-chan v7pushaction.PushState,
	eventStream <-chan v7pushaction.Event,
	warningsStream <-chan v7pushaction.Warnings,
	errorStream <-chan error,
) (v7pushaction.PushState, error) {
	var stateClosed, eventClosed, warningsClosed, errClosed, complete bool
	var updateState v7pushaction.PushState

	for {
		select {
		case state, ok := <-stateStream:
			if !ok {
				if !stateClosed {
					log.Debug("processing config stream closed")
				}
				stateClosed = true
				break
			}
			updateState = state
		case event, ok := <-eventStream:
			if !ok {
				if !eventClosed {
					log.Debug("processing event stream closed")
				}
				eventClosed = true
				break
			}
			complete = cmd.processEvent(appName, event)
		case warnings, ok := <-warningsStream:
			if !ok {
				if !warningsClosed {
					log.Debug("processing warnings stream closed")
				}
				warningsClosed = true
				break
			}
			cmd.UI.DisplayWarnings(warnings)
		case err, ok := <-errorStream:
			if !ok {
				if !errClosed {
					log.Debug("processing error stream closed")
				}
				errClosed = true
				break
			}
			return v7pushaction.PushState{}, err
		}

		if stateClosed && eventClosed && warningsClosed && complete {
			break
		}
	}

	return updateState, nil
}

func (cmd PushCommand) processEvent(appName string, event v7pushaction.Event) bool {
	log.Infoln("received apply event:", event)

	switch event {
	case v7pushaction.SkippingApplicationCreation:
		cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case v7pushaction.CreatedApplication:
		cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case v7pushaction.CreatingAndMappingRoutes:
		cmd.UI.DisplayText("Mapping routes...")
	case v7pushaction.CreatingArchive:
		cmd.UI.DisplayText("Packaging files to upload...")
	case v7pushaction.UploadingApplicationWithArchive:
		cmd.UI.DisplayText("Uploading files...")
		log.Debug("starting progress bar")
		cmd.ProgressBar.Ready()
	case v7pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case v7pushaction.UploadWithArchiveComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.StartingStaging:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Staging app and tracing logs...")
		logStream, errStream, warnings, _ := cmd.VersionActor.GetStreamingLogsForApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
		cmd.UI.DisplayWarnings(warnings)
		go cmd.getLogs(logStream, errStream)
	case v7pushaction.StagingComplete:
		cmd.NOAAClient.Close()
	case v7pushaction.Complete:
		return true
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false
}

func (cmd PushCommand) getLogs(logStream <-chan *v7action.LogMessage, errStream <-chan error) {
	for {
		select {
		case logMessage, open := <-logStream:
			if !open {
				return
			}
			if logMessage.Staging() {
				cmd.UI.DisplayLogMessage(logMessage, false)
			}
		case err, open := <-errStream:
			if !open {
				return
			}
			_, ok := err.(actionerror.NOAATimeoutError)
			if ok {
				cmd.UI.DisplayWarning("timeout connecting to log server, no log will be shown")
			}
			cmd.UI.DisplayWarning(err.Error())
		}
	}
}

func (cmd PushCommand) GetFlagOverrides() (v7pushaction.FlagOverrides, error) {
	return v7pushaction.FlagOverrides{
		Buildpacks:      cmd.Buildpacks,
		HealthCheckType: cmd.HealthCheckType.Type,
		Memory:          cmd.Memory.NullUint64,
		ProvidedAppPath: string(cmd.AppPath),
		StartCommand:    cmd.StartCommand.FilteredString,
	}, nil
}
