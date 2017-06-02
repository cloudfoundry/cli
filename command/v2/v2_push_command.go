package v2

import (
	"os"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/progressbar"
	"code.cloudfoundry.org/cli/util/ui"
	log "github.com/Sirupsen/logrus"
	"github.com/cloudfoundry/noaa/consumer"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	pushaction.ProgressBar
	Complete()
	Ready()
}

//go:generate counterfeiter . V2PushActor

type V2PushActor interface {
	Apply(config pushaction.ApplicationConfig, progressBar pushaction.ProgressBar) (<-chan pushaction.ApplicationConfig, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error)
	ConvertToApplicationConfigs(orgGUID string, spaceGUID string, apps []manifest.Application) ([]pushaction.ApplicationConfig, pushaction.Warnings, error)
	MergeAndValidateSettingsAndManifests(cmdSettings pushaction.CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error)
}

type V2PushCommand struct {
	OptionalArgs         flag.AppName                `positional-args:"yes"`
	BuildpackName        string                      `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	StartupCommand       string                      `short:"c" description:"Startup command, set to null to reset to default start command"`
	Domain               string                      `short:"d" description:"Domain (e.g. example.com)"`
	DockerImage          string                      `long:"docker-image" short:"o" description:"Docker-image to be used (e.g. user/docker-image-name)"`
	PathToManifest       flag.PathWithExistenceCheck `short:"f" description:"Path to manifest"`
	HealthCheckType      flag.HealthCheckType        `long:"health-check-type" short:"u" description:"Application health check type (Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/')"`
	Hostname             string                      `long:"hostname" short:"n" description:"Hostname (e.g. my-subdomain)"`
	NumInstances         int                         `short:"i" description:"Number of instances"`
	DiskLimit            string                      `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit          string                      `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoHostname           bool                        `long:"no-hostname" description:"Map the root domain to this app"`
	NoManifest           bool                        `long:"no-manifest" description:"Ignore manifest file"`
	NoRoute              bool                        `long:"no-route" description:"Do not map a route to this app and remove routes from previous pushes of this app"`
	NoStart              bool                        `long:"no-start" description:"Do not start an app after pushing"`
	DirectoryPath        flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	RandomRoute          bool                        `long:"random-route" description:"Create a random route for this app"`
	RoutePath            string                      `long:"route-path" description:"Path for the route"`
	Stack                string                      `short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	ApplicationStartTime int                         `short:"t" description:"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app"`

	usage               interface{} `usage:"Push a single app (with or without a manifest):\n   CF_NAME v2-push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND] [-d DOMAIN] [-f MANIFEST_PATH] [--docker-image DOCKER_IMAGE]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [--hostname HOST] [-p PATH] [-s STACK] [-t TIMEOUT] [-u (process | port | http)] [--route-path ROUTE_PATH]\n   [--no-hostname] [--no-manifest] [--no-route] [--no-start] [--random-route]\n\n   Push multiple apps with a manifest:\n   cf v2-push [-f MANIFEST_PATH]"`
	envCFStagingTimeout interface{} `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{} `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
	relatedCommands     interface{} `related_commands:"apps, create-app-manifest, logs, ssh, start"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V2PushActor
	ProgressBar ProgressBar

	StartActor StartActor
	NOAAClient *consumer.Consumer
}

func (cmd *V2PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	v2Actor := v2action.NewActor(ccClient, uaaClient)
	cmd.StartActor = v2Actor
	cmd.Actor = pushaction.NewActor(v2Actor)

	cmd.NOAAClient = shared.NewNOAAClient(ccClient.DopplerEndpoint(), config, uaaClient, ui)

	cmd.ProgressBar = progressbar.NewProgressBar()
	return nil
}

func (cmd V2PushCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	log.Info("collating flags")
	cliSettings, err := cmd.GetCommandLineSettings()
	if err != nil {
		log.Errorln("reading flags:", err)
		return shared.HandleError(err)
	}

	//TODO: Read in manifest
	log.Info("merging manifest and command flags")
	manifestApplications, err := cmd.Actor.MergeAndValidateSettingsAndManifests(cliSettings, nil)
	if err != nil {
		log.Errorln("merging manifest:", err)
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText("Getting app info...")

	log.Info("converting manifests to ApplicationConfigs")
	appConfigs, warnings, err := cmd.Actor.ConvertToApplicationConfigs(
		cmd.Config.TargetedOrganization().GUID,
		cmd.Config.TargetedSpace().GUID,
		manifestApplications,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		log.Errorln("converting manifest:", err)
		return shared.HandleError(err)
	}

	for _, appConfig := range appConfigs {
		log.Infoln("starting create/update:", appConfig.DesiredApplication.Name)
		cmd.displayChanges(appConfig)
		configStream, eventStream, warningsStream, errorStream := cmd.Actor.Apply(appConfig, cmd.ProgressBar)
		updatedConfig, err := cmd.processApplyStreams(user, appConfig, configStream, eventStream, warningsStream, errorStream)
		if err != nil {
			return shared.HandleError(err)
		}

		if appConfig.CurrentApplication.Started() {
			cmd.UI.DisplayText("Stopping app...")
		}

		messages, logErrs, appStarting, apiWarnings, errs := cmd.StartActor.RestartApplication(updatedConfig.CurrentApplication, cmd.NOAAClient, cmd.Config)
		err = shared.PollStart(cmd.UI, cmd.Config, messages, logErrs, appStarting, apiWarnings, errs)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		appSummary, warnings, err := cmd.StartActor.GetApplicationSummaryByNameAndSpace(appConfig.DesiredApplication.Name, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return shared.HandleError(err)
		}

		shared.DisplayAppSummary(cmd.UI, appSummary, true)
	}

	return nil
}

func (cmd V2PushCommand) GetCommandLineSettings() (pushaction.CommandLineSettings, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return pushaction.CommandLineSettings{}, err
	}

	config := pushaction.CommandLineSettings{
		Name:             cmd.OptionalArgs.AppName,
		CurrentDirectory: pwd,
	}

	log.Debugf("%#v", config)
	return config, nil
}

func (cmd V2PushCommand) displayChanges(appConfig pushaction.ApplicationConfig) error {
	if appConfig.CreatingApplication() {
		cmd.UI.DisplayText("Creating app with these attributes...")
	} else {
		cmd.UI.DisplayText("Updating app with these attributes...")
	}

	var currentRoutes []string
	for _, route := range appConfig.CurrentRoutes {
		currentRoutes = append(currentRoutes, route.String())
	}

	var desiredRotues []string
	for _, route := range appConfig.DesiredRoutes {
		desiredRotues = append(desiredRotues, route.String())
	}

	err := cmd.UI.DisplayChangesForPush([]ui.Change{
		{
			Header:       "name:",
			CurrentValue: appConfig.CurrentApplication.Name,
			NewValue:     appConfig.DesiredApplication.Name,
		},
		{
			Header:       "path:",
			CurrentValue: appConfig.Path,
			NewValue:     appConfig.Path,
		},
		{
			Header:       "routes:",
			CurrentValue: currentRoutes,
			NewValue:     desiredRotues,
		},
	})

	if err != nil {
		log.Errorln("display changes:", err)
		return shared.HandleError(err)
	}

	cmd.UI.DisplayNewline()
	return nil
}

func (cmd V2PushCommand) processApplyStreams(
	user configv3.User,
	appConfig pushaction.ApplicationConfig,
	configStream <-chan pushaction.ApplicationConfig,
	eventStream <-chan pushaction.Event,
	warningsStream <-chan pushaction.Warnings,
	errorStream <-chan error,
) (pushaction.ApplicationConfig, error) {
	var configClosed, eventClosed, warningsClosed, complete bool
	var updatedConfig pushaction.ApplicationConfig

	for {
		select {
		case config, ok := <-configStream:
			if !ok {
				log.Debug("processing config stream closed")
				configClosed = true
				break
			}
			updatedConfig = config
			log.Debugf("updated config received: %#v", updatedConfig)
		case event, ok := <-eventStream:
			if !ok {
				log.Debug("processing event stream closed")
				eventClosed = true
				break
			}
			complete = cmd.processEvent(user, appConfig, event)
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
			return pushaction.ApplicationConfig{}, err
		}

		if configClosed && eventClosed && warningsClosed && complete {
			log.Debug("breaking apply display loop")
			break
		}
	}

	return updatedConfig, nil
}

func (cmd V2PushCommand) processEvent(user configv3.User, appConfig pushaction.ApplicationConfig, event pushaction.Event) bool {
	log.Infoln("received apply event:", event)

	switch event {
	case pushaction.ConfiguringRoutes:
		cmd.UI.DisplayText("Mapping routes...")
	case pushaction.CreatingArchive:
		cmd.UI.DisplayText("Packaging files to upload...")
	case pushaction.UploadingApplication:
		cmd.UI.DisplayText("Uploading files...")
		log.Debug("starting progress bar")
		cmd.ProgressBar.Ready()
	case pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case pushaction.UploadComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case pushaction.Complete:
		return true
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false
}
