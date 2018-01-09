package v2

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/manifest"
	"code.cloudfoundry.org/cli/util/progressbar"
	"github.com/cloudfoundry/noaa/consumer"
	log "github.com/sirupsen/logrus"
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
	ConvertToApplicationConfigs(orgGUID string, spaceGUID string, noStart bool, apps []manifest.Application) ([]pushaction.ApplicationConfig, pushaction.Warnings, error)
	MergeAndValidateSettingsAndManifests(cmdSettings pushaction.CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error)
	ReadManifest(pathToManifest string) ([]manifest.Application, error)
}

type V2PushCommand struct {
	OptionalArgs        flag.OptionalAppName        `positional-args:"yes"`
	Buildpack           flag.Buildpack              `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	Command             flag.Command                `short:"c" description:"Startup command, set to null to reset to default start command"`
	Domain              string                      `short:"d" description:"Domain (e.g. example.com)"`
	DockerImage         flag.DockerImage            `long:"docker-image" short:"o" description:"Docker-image to be used (e.g. user/docker-image-name)"`
	DockerUsername      string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	PathToManifest      flag.PathWithExistenceCheck `short:"f" description:"Path to manifest"`
	HealthCheckType     flag.HealthCheckType        `long:"health-check-type" short:"u" description:"Application health check type (Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/')"`
	Hostname            string                      `long:"hostname" short:"n" description:"Hostname (e.g. my-subdomain)"`
	Instances           flag.Instances              `short:"i" description:"Number of instances"`
	DiskQuota           flag.Megabytes              `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	Memory              flag.Megabytes              `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoHostname          bool                        `long:"no-hostname" description:"Map the root domain to this app"`
	NoManifest          bool                        `long:"no-manifest" description:"Ignore manifest file"`
	NoRoute             bool                        `long:"no-route" description:"Do not map a route to this app and remove routes from previous pushes of this app"`
	NoStart             bool                        `long:"no-start" description:"Do not start an app after pushing"`
	AppPath             flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	RandomRoute         bool                        `long:"random-route" description:"Create a random route for this app"`
	RoutePath           flag.RoutePath              `long:"route-path" description:"Path for the route"`
	StackName           string                      `short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	HealthCheckTimeout  int                         `short:"t" description:"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app"`
	envCFStagingTimeout interface{}                 `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}                 `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
	dockerPassword      interface{}                 `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`

	usage           interface{} `usage:"cf push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\n   [--no-route | --random-route | --hostname HOST | --no-hostname] [-d DOMAIN] [--route-path ROUTE_PATH]\n\n   cf push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME]\n   [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\n   [--no-route | --random-route | --hostname HOST | --no-hostname] [-d DOMAIN] [--route-path ROUTE_PATH]\n\n   cf push -f MANIFEST_WITH_MULTIPLE_APPS_PATH [APP_NAME] [--no-start]"`
	relatedCommands interface{} `related_commands:"apps, create-app-manifest, logs, ssh, start"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V2PushActor
	ProgressBar ProgressBar

	RestartActor RestartActor
	NOAAClient   *consumer.Consumer
}

func (cmd *V2PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	v2Actor := v2action.NewActor(ccClient, uaaClient, config)
	cmd.RestartActor = v2Actor
	cmd.Actor = pushaction.NewActor(v2Actor, sharedActor)
	cmd.SharedActor = sharedActor
	cmd.NOAAClient = shared.NewNOAAClient(ccClient.DopplerEndpoint(), config, uaaClient, ui)

	cmd.ProgressBar = progressbar.NewProgressBar()
	return nil
}

func (cmd V2PushCommand) Execute(args []string) error {
	// cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	log.Info("collating flags")
	cliSettings, err := cmd.GetCommandLineSettings()
	if err != nil {
		log.Errorln("reading flags:", err)
		return err
	}

	log.Info("checking manifest")
	rawApps, err := cmd.findAndReadManifestWithFlavorText(cliSettings)
	if err != nil {
		log.Errorln("reading manifest:", err)
		return err
	}

	log.Info("merging manifest and command flags")
	manifestApplications, err := cmd.Actor.MergeAndValidateSettingsAndManifests(cliSettings, rawApps)
	if err != nil {
		log.Errorln("merging manifest:", err)
		return err
	}

	cmd.UI.DisplayText("Getting app info...")

	log.Info("converting manifests to ApplicationConfigs")
	appConfigs, warnings, err := cmd.Actor.ConvertToApplicationConfigs(
		cmd.Config.TargetedOrganization().GUID,
		cmd.Config.TargetedSpace().GUID,
		cmd.NoStart,
		manifestApplications,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		log.Errorln("converting manifest:", err)
		return err
	}

	for _, appConfig := range appConfigs {
		if appConfig.CreatingApplication() {
			cmd.UI.DisplayText("Creating app with these attributes...")
		} else {
			cmd.UI.DisplayText("Updating app with these attributes...")
		}
		log.Infoln("starting create/update:", appConfig.DesiredApplication.Name)
		changes := shared.GetApplicationChanges(appConfig)
		err := cmd.UI.DisplayChangesForPush(changes)
		if err != nil {
			log.Errorln("display changes:", err)
			return err
		}
		cmd.UI.DisplayNewline()
	}

	for appNumber, appConfig := range appConfigs {
		if appConfig.CreatingApplication() {
			cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}}...", map[string]interface{}{
				"AppName": appConfig.DesiredApplication.Name,
			})
		} else {
			cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}}...", map[string]interface{}{
				"AppName": appConfig.DesiredApplication.Name,
			})
		}

		configStream, eventStream, warningsStream, errorStream := cmd.Actor.Apply(appConfig, cmd.ProgressBar)
		updatedConfig, err := cmd.processApplyStreams(user, appConfig, configStream, eventStream, warningsStream, errorStream)
		if err != nil {
			log.Errorln("process apply stream:", err)
			return err
		}

		if !cmd.NoStart {
			messages, logErrs, appState, apiWarnings, errs := cmd.RestartActor.RestartApplication(updatedConfig.CurrentApplication.Application, cmd.NOAAClient, cmd.Config)
			err = shared.PollStart(cmd.UI, cmd.Config, messages, logErrs, appState, apiWarnings, errs)
			if err != nil {
				return err
			}
		}

		cmd.UI.DisplayNewline()
		appSummary, warnings, err := cmd.RestartActor.GetApplicationSummaryByNameAndSpace(appConfig.DesiredApplication.Name, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		shared.DisplayAppSummary(cmd.UI, appSummary, true)

		if appNumber+1 <= len(appConfigs) {
			cmd.UI.DisplayNewline()
		}
	}

	return nil
}

// GetCommandLineSettings generates a push CommandLineSettings object from the
// command's command line flags. It also validates those settings, preventing
// contradictory flags.
func (cmd V2PushCommand) GetCommandLineSettings() (pushaction.CommandLineSettings, error) {
	err := cmd.validateArgs()
	if err != nil {
		return pushaction.CommandLineSettings{}, err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return pushaction.CommandLineSettings{}, err
	}

	dockerPassword := cmd.Config.DockerPassword()
	if dockerPassword != "" {
		cmd.UI.DisplayText("Using docker repository password from environment variable CF_DOCKER_PASSWORD.")
	} else if cmd.DockerUsername != "" {
		cmd.UI.DisplayText("Environment variable CF_DOCKER_PASSWORD not set.")
		dockerPassword, err = cmd.UI.DisplayPasswordPrompt("Docker password")
		if err != nil {
			return pushaction.CommandLineSettings{}, err
		}
	}

	config := pushaction.CommandLineSettings{
		Buildpack:            cmd.Buildpack.FilteredString, // -b
		Command:              cmd.Command.FilteredString,   // -c
		CurrentDirectory:     pwd,
		DefaultRouteDomain:   cmd.Domain,               // -d
		DefaultRouteHostname: cmd.Hostname,             // -n/--hostname
		DiskQuota:            cmd.DiskQuota.Value,      // -k
		DockerImage:          cmd.DockerImage.Path,     // -o
		DockerPassword:       dockerPassword,           // ENV - CF_DOCKER_PASSWORD
		DockerUsername:       cmd.DockerUsername,       // --docker-username
		HealthCheckTimeout:   cmd.HealthCheckTimeout,   // -t
		HealthCheckType:      cmd.HealthCheckType.Type, // -u/--health-check-type
		Instances:            cmd.Instances.NullInt,    // -i
		Memory:               cmd.Memory.Value,         // -m
		Name:                 cmd.OptionalArgs.AppName, // arg
		NoHostname:           cmd.NoHostname,           // --no-hostname
		NoRoute:              cmd.NoRoute,              // --no-route
		ProvidedAppPath:      string(cmd.AppPath),      // -p
		RandomRoute:          cmd.RandomRoute,          // --random-route
		RoutePath:            cmd.RoutePath.Path,       // --route-path
		StackName:            cmd.StackName,            // -s
	}

	log.Debugln("Command Line Settings:", config)
	return config, nil
}

func (cmd V2PushCommand) findAndReadManifestWithFlavorText(settings pushaction.CommandLineSettings) ([]manifest.Application, error) {
	var pathToManifest string

	switch {
	case cmd.NoManifest:
		log.Debug("skipping reading of manifest")
	case cmd.PathToManifest != "":
		log.WithField("file", cmd.PathToManifest).Debug("using specified manifest file")
		pathToManifest = string(cmd.PathToManifest)

		fileInfo, err := os.Stat(pathToManifest)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			manifestPaths := []string{
				filepath.Join(pathToManifest, "manifest.yml"),
				filepath.Join(pathToManifest, "manifest.yaml"),
			}
			for _, manifestPath := range manifestPaths {
				if _, err = os.Stat(manifestPath); err == nil {
					pathToManifest = manifestPath
					break
				}
			}
		}

		if err != nil {
			return nil, translatableerror.ManifestFileNotFoundInDirectoryError{
				PathToManifest: pathToManifest,
			}
		}
	default:
		log.Debug("searching for manifest file")
		pathToManifest = filepath.Join(settings.CurrentDirectory, "manifest.yml")
		if _, err := os.Stat(pathToManifest); os.IsNotExist(err) {
			log.WithField("pathToManifest", pathToManifest).Debug("could not find")

			// While this is unlikely to be used, it is kept for backwards
			// compatibility.
			pathToManifest = filepath.Join(settings.CurrentDirectory, "manifest.yaml")
			if _, err := os.Stat(pathToManifest); os.IsNotExist(err) {
				log.WithField("pathToManifest", pathToManifest).Debug("could not find")
				pathToManifest = ""
			}
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return nil, err
	}

	if pathToManifest == "" {
		cmd.UI.DisplayTextWithFlavor("Pushing app {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   settings.Name,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
		return nil, nil
	}

	cmd.UI.DisplayTextWithFlavor("Pushing from manifest to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	log.WithField("pathToManifest", pathToManifest).Info("reading manifest")
	cmd.UI.DisplayText("Using manifest file {{.Path}}", map[string]interface{}{
		"Path": pathToManifest,
	})
	return cmd.Actor.ReadManifest(pathToManifest)
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
	case pushaction.CreatingAndMappingRoutes:
		cmd.UI.DisplayText("Mapping routes...")
	case pushaction.UnmappingRoutes:
		cmd.UI.DisplayText("Unmapping routes...")
	case pushaction.ConfiguringServices:
		cmd.UI.DisplayText("Binding services...")
	case pushaction.ResourceMatching:
		cmd.UI.DisplayText("Comparing local files to remote cache...")
	case pushaction.CreatingArchive:
		cmd.UI.DisplayText("Packaging files to upload...")
	case pushaction.UploadingApplication:
		cmd.UI.DisplayText("All files found in remote cache; nothing to upload.")
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case pushaction.UploadingApplicationWithArchive:
		cmd.UI.DisplayText("Uploading files...")
		log.Debug("starting progress bar")
		cmd.ProgressBar.Ready()
	case pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case pushaction.UploadWithArchiveComplete:
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

func (cmd V2PushCommand) validateArgs() error {
	switch {
	case cmd.DockerImage.Path != "" && cmd.AppPath != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--docker-image, -o", "-p"},
		}
	case cmd.DockerImage.Path != "" && cmd.Buildpack.IsSet:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-b", "--docker-image, -o"},
		}
	case cmd.DockerUsername != "" && cmd.DockerImage.Path == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--docker-image, -o",
			Arg2: "--docker-username",
		}
	case cmd.Domain != "" && cmd.NoRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-d", "--no-route"},
		}
	case cmd.Hostname != "" && cmd.NoHostname:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--hostname", "-n", "--no-hostname"},
		}
	case cmd.Hostname != "" && cmd.NoRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--hostname", "-n", "--no-route"},
		}
	case cmd.NoHostname && cmd.NoRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--no-hostname", "--no-route"},
		}
	case cmd.PathToManifest != "" && cmd.NoManifest:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-f", "--no-manifest"},
		}
	case cmd.RandomRoute && cmd.Hostname != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--hostname", "-n", "--random-route"},
		}
	case cmd.RandomRoute && cmd.NoHostname:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--no-hostname", "--random-route"},
		}
	case cmd.RandomRoute && cmd.NoRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--no-route", "--random-route"},
		}
	case cmd.RandomRoute && cmd.RoutePath.Path != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--random-route", "--route-path"},
		}
	case cmd.RoutePath.Path != "" && cmd.NoRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--route-path", "--no-route"},
		}
	}

	return nil
}
