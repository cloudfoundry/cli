package v3

import (
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/progressbar"

	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	pushaction.ProgressBar
	Complete()
	Ready()
}

//go:generate counterfeiter . V3PushActor

type V3PushActor interface {
	Actualize(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error)
	GeneratePushState(setting pushaction.CommandLineSettings, spaceGUID string) ([]pushaction.PushState, pushaction.Warnings, error)
}

//go:generate counterfeiter . V3PushVersionActor

type V3PushVersionActor interface {
	CloudControllerAPIVersion() string
}

type V3PushCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	Buildpacks   []string     `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	// Command flag.Command
	// Domain string
	DockerImage    flag.DockerImage `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername string           `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	// DropletPath flag.PathWithExistenceCheck
	// PathToManifest flag.PathWithExistenceCheck
	// HealthCheckType flag.HealthCheckType
	// Hostname string
	// Instances flag.Instances
	// DiskQuota           flag.Megabytes
	// Memory              flag.Megabytes
	// NoHostname          bool
	// NoManifest          bool
	NoRoute bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	// RandomRoute         bool
	// RoutePath           flag.RoutePath
	// StackName           string
	// VarsFilePaths []flag.PathWithExistenceCheck
	// Vars []template.VarKV
	// HealthCheckTimeout int
	dockerPassword      interface{} `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage               interface{} `usage:"cf v3-push APP_NAME [-b BUILDPACK]... [-p APP_PATH] [--no-route] [--no-start]\n   cf v3-push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME] [--no-route] [--no-start]"`
	envCFStagingTimeout interface{} `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{} `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                  command.UI
	Config              command.Config
	NOAAClient          v3action.NOAAClient
	Actor               V3PushActor
	VersionActor        V3PushVersionActor
	SharedActor         command.SharedActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
	PackageDisplayer    shared.PackageDisplayer
	ProgressBar         ProgressBar

	OriginalActor       OriginalV3PushActor
	OriginalV2PushActor OriginalV2PushActor
}

func (cmd *V3PushCommand) Setup(config command.Config, ui command.UI) error {
	if !config.Experimental() {
		return cmd.OriginalSetup(config, ui)
	}

	cmd.Config = config
	cmd.UI = ui
	cmd.ProgressBar = progressbar.NewProgressBar()

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	v3Actor := v3action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.VersionActor = v3Actor

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)
	cmd.Actor = pushaction.NewActor(v2Actor, v3Actor, sharedActor)

	return nil
}

func (cmd V3PushCommand) Execute(args []string) error {
	if !cmd.Config.Experimental() {
		return cmd.OriginalExecute(args)
	}

	err := command.MinimumAPIVersionCheck(cmd.VersionActor.CloudControllerAPIVersion(), ccversion.MinVersionV3)
	if err != nil {
		return err
	}

	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Getting app info...")

	log.Info("generating the app state")
	pushState, warnings, err := cmd.Actor.GeneratePushState(pushaction.CommandLineSettings{
		Buildpacks:      cmd.Buildpacks,
		Name:            cmd.RequiredArgs.AppName,
		ProvidedAppPath: string(cmd.AppPath),
	}, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	log.WithField("number of states", len(pushState)).Debug("completed generating state")

	for _, state := range pushState {
		log.WithField("app_name", state.Application.Name).Info("actualizing")
		stateStream, eventStream, warningsStream, errorStream := cmd.Actor.Actualize(state, cmd.ProgressBar)
		_, err = cmd.processApplyStreams(state.Application.Name, stateStream, eventStream, warningsStream, errorStream)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd V3PushCommand) processApplyStreams(
	appName string,
	stateStream <-chan pushaction.PushState,
	eventStream <-chan pushaction.Event,
	warningsStream <-chan pushaction.Warnings,
	errorStream <-chan error,
) (pushaction.PushState, error) {
	var stateClosed, eventClosed, warningsClosed, complete bool
	var updateState pushaction.PushState

	for {
		select {
		case state, ok := <-stateStream:
			if !ok {
				log.Debug("processing config stream closed")
				stateClosed = true
				break
			}
			updateState = state
			log.Debugf("updated config received: %#v", updateState)
		case event, ok := <-eventStream:
			if !ok {
				log.Debug("processing event stream closed")
				eventClosed = true
				break
			}
			complete = cmd.processEvent(appName, event)
		case warnings, ok := <-warningsStream:
			if !ok {
				log.Debug("processing warnings stream closed")
				warningsClosed = true
				break
			}
			cmd.UI.DisplayWarnings(warnings)
		case err, ok := <-errorStream:
			if !ok {
				log.Debug("processing error stream closed")
				warningsClosed = true
				break
			}
			return pushaction.PushState{}, err
		}

		if stateClosed && eventClosed && warningsClosed && complete {
			log.Debug("breaking apply display loop")
			break
		}
	}

	return updateState, nil
}

func (cmd V3PushCommand) processEvent(appName string, event pushaction.Event) bool {
	log.Infoln("received apply event:", event)

	switch event {
	case pushaction.CreatedApplication:
		cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case pushaction.SkipingApplicationCreation:
		cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case pushaction.Complete:
		return true
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false
}
