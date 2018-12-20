package v7

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
	Conceptualize(appName string, spaceGUID string, orgGUID string, currentDir string, flagOverrides v7pushaction.FlagOverrides, manifest []byte) ([]v7pushaction.PushState, v7pushaction.Warnings, error)
}

//go:generate counterfeiter . V7ActorForPush

type V7ActorForPush interface {
	AppActor
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	RestartApplication(appGUID string) (v7action.Warnings, error)
}

type PushCommand struct {
	RequiredArgs        flag.AppName                `positional-args:"yes"`
	Buildpacks          []string                    `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	DockerImage         flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername      string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	HealthCheckType     flag.HealthCheckType        `long:"health-check-type" short:"u" description:"Application health check type: 'port' (default), 'process', 'http' (implies endpoint '/')"`
	Instances           flag.Instances              `long:"instances" short:"i" description:"Number of instances"`
	PathToManifest      flag.PathWithExistenceCheck `long:"manifest" short:"f" description:"Path to manifest"`
	Memory              flag.Megabytes              `long:"memory" short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoRoute             bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart             bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath             flag.PathWithExistenceCheck `long:"path" short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
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

	return err
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

	manifest, err := cmd.readManifest()
	if err != nil {
		return err
	}

	log.Info("generating the app state")
	pushState, warnings, err := cmd.Actor.Conceptualize(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		cmd.PWD,
		overrides,
		manifest,
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

		anyProcessCrashed := false
		if !cmd.NoStart {
			cmd.UI.DisplayNewline()
			cmd.UI.DisplayText("Waiting for app to start...")
			warnings, restartErr := cmd.VersionActor.RestartApplication(updatedState.Application.GUID)
			cmd.UI.DisplayWarnings(warnings)

			if restartErr != nil {
				if _, ok := restartErr.(actionerror.StartupTimeoutError); ok {
					return translatableerror.StartupTimeoutError{
						AppName:    cmd.RequiredArgs.AppName,
						BinaryName: cmd.Config.BinaryName(),
					}
				} else if _, ok := restartErr.(actionerror.AllInstancesCrashedError); ok {
					anyProcessCrashed = true
				} else {
					return restartErr
				}
			}
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

		if anyProcessCrashed {
			return translatableerror.ApplicationUnableToStartError{
				AppName:    cmd.RequiredArgs.AppName,
				BinaryName: cmd.Config.BinaryName(),
			}
		}
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
	case v7pushaction.CreatingApplication:
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
	case v7pushaction.StoppingApplication:
		cmd.UI.DisplayText("Stopping Application...")
	case v7pushaction.StoppingApplicationComplete:
		cmd.UI.DisplayText("Application Stopped")
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

func (cmd PushCommand) readManifest() ([]byte, error) {
	log.Info("reading manifest if exists")

	if len(cmd.PathToManifest) != 0 {
		log.WithField("manifestPath", cmd.PathToManifest).Debug("reading '-f' provided manifest")
		return ioutil.ReadFile(string(cmd.PathToManifest))
	}

	manifestPath := filepath.Join(cmd.PWD, "manifest.yml")
	log.WithField("manifestPath", manifestPath).Debug("path to manifest")

	manifest, err := ioutil.ReadFile(manifestPath)
	if err != nil && !os.IsNotExist(err) {
		log.Errorln("reading manifest:", err)
		return nil, err
	} else if os.IsNotExist(err) {
		log.Debug("no manifest exists")
	}
	return manifest, nil
}

func (cmd PushCommand) GetFlagOverrides() (v7pushaction.FlagOverrides, error) {
	var dockerPassword string
	if cmd.DockerUsername != "" {
		if cmd.Config.DockerPassword() == "" {
			var err error
			cmd.UI.DisplayText("Environment variable CF_DOCKER_PASSWORD not set.")
			dockerPassword, err = cmd.UI.DisplayPasswordPrompt("Docker password")
			if err != nil {
				return v7pushaction.FlagOverrides{}, err
			}
		} else {
			cmd.UI.DisplayText("Using docker repository password from environment variable CF_DOCKER_PASSWORD.")
			dockerPassword = cmd.Config.DockerPassword()
		}
	}

	return v7pushaction.FlagOverrides{
		Buildpacks:        cmd.Buildpacks,
		DockerImage:       cmd.DockerImage.Path,
		DockerPassword:    dockerPassword,
		DockerUsername:    cmd.DockerUsername,
		HealthCheckType:   cmd.HealthCheckType.Type,
		Instances:         cmd.Instances.NullInt,
		Memory:            cmd.Memory.NullUint64,
		ProvidedAppPath:   string(cmd.AppPath),
		SkipRouteCreation: cmd.NoRoute,
		StartCommand:      cmd.StartCommand.FilteredString,
		NoStart:           cmd.NoStart,
	}, nil
}
