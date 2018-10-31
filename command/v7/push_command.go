package v7

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/progressbar"

	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	pushaction.ProgressBar
	Complete()
	Ready()
}

//go:generate counterfeiter . PushActor

type PushActor interface {
	Actualize(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error)
	Conceptualize(setting pushaction.CommandLineSettings, spaceGUID string) ([]pushaction.PushState, pushaction.Warnings, error)
}

//go:generate counterfeiter . PushVersionActor

type PushVersionActor interface {
	CloudControllerAPIVersion() string
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error)
	PollStart(appGUID string, warningsChannel chan<- v3action.Warnings) error
	RestartApplication(appGUID string) (v3action.Warnings, error)
}

type PushCommand struct {
	RequiredArgs        flag.AppName                `positional-args:"yes"`
	Buildpacks          []string                    `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	DockerImage         flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername      string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	NoRoute             bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart             bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath             flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	dockerPassword      interface{}                 `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage               interface{}                 `usage:"CF_NAME push APP_NAME [-b BUILDPACK]... [-p APP_PATH] [--no-route] [--no-start]\n   CF_NAME push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME] [--no-route] [--no-start]"`
	envCFStagingTimeout interface{}                 `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}                 `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                  command.UI
	Config              command.Config
	NOAAClient          v3action.NOAAClient
	Actor               PushActor
	VersionActor        PushVersionActor
	SharedActor         command.SharedActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
	PackageDisplayer    shared.PackageDisplayer
	ProgressBar         ProgressBar
}

func (cmd *PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.ProgressBar = progressbar.NewProgressBar()

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		return err
	}
	v3Actor := v3action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.VersionActor = v3Actor

	ccClientV2, uaaClientV2, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)
	cmd.Actor = pushaction.NewActor(v2Actor, v3Actor, sharedActor)

	cmd.NOAAClient = shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

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

	cliSettings, err := cmd.GetCommandLineSettings()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Pushing app {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cliSettings.Name,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	cmd.UI.DisplayText("Getting app info...")

	log.Info("generating the app state")
	pushState, warnings, err := cmd.Actor.Conceptualize(cliSettings, cmd.Config.TargetedSpace().GUID)
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

		pollWarnings := make(chan v3action.Warnings)
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
	}

	return nil
}

func (cmd PushCommand) processApplyStreams(
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
			break
		}
	}

	return updateState, nil
}

func (cmd PushCommand) processEvent(appName string, event pushaction.Event) bool {
	log.Infoln("received apply event:", event)

	switch event {
	case pushaction.SkippingApplicationCreation:
		cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case pushaction.CreatedApplication:
		cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case pushaction.CreatingArchive:
		cmd.UI.DisplayTextWithFlavor("Packaging files to upload...")
	case pushaction.UploadingApplicationWithArchive:
		cmd.UI.DisplayTextWithFlavor("Uploading files...")
		log.Debug("starting progress bar")
		cmd.ProgressBar.Ready()
	case pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case pushaction.UploadWithArchiveComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case pushaction.StartingStaging:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Staging app and tracing logs...")
		logStream, errStream, warnings, _ := cmd.VersionActor.GetStreamingLogsForApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
		cmd.UI.DisplayWarnings(warnings)
		go cmd.getLogs(logStream, errStream)
	case pushaction.StagingComplete:
		cmd.NOAAClient.Close()
	case pushaction.Complete:
		return true
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false
}

func (cmd PushCommand) getLogs(logStream <-chan *v3action.LogMessage, errStream <-chan error) {
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

func (cmd PushCommand) GetCommandLineSettings() (pushaction.CommandLineSettings, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return pushaction.CommandLineSettings{}, err
	}
	return pushaction.CommandLineSettings{
		Buildpacks:       cmd.Buildpacks,
		CurrentDirectory: pwd,
		Name:             cmd.RequiredArgs.AppName,
		ProvidedAppPath:  string(cmd.AppPath),
	}, nil
}
