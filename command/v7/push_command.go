package v7

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

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
	"code.cloudfoundry.org/cli/util/manifestparser"
	"code.cloudfoundry.org/cli/util/progressbar"

	"github.com/cloudfoundry/bosh-cli/director/template"
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
	CreatePushPlans(appNameArg string, spaceGUID string, orgGUID string, parser v7pushaction.ManifestParser, overrides v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, error)
	// Prepare the space by creating needed apps/applying the manifest
	PrepareSpace(pushPlans []v7pushaction.PushPlan, parser v7pushaction.ManifestParser) (<-chan []v7pushaction.PushPlan, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error)
	// UpdateApplicationSettings figures out the state of the world.
	UpdateApplicationSettings(pushPlans []v7pushaction.PushPlan) ([]v7pushaction.PushPlan, v7pushaction.Warnings, error)
	// Actualize applies any necessary changes.
	Actualize(plan v7pushaction.PushPlan, progressBar v7pushaction.ProgressBar) (<-chan v7pushaction.PushPlan, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error)
}

//go:generate counterfeiter . V7ActorForPush

type V7ActorForPush interface {
	AppActor
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	RestartApplication(appGUID string) (v7action.Warnings, error)
}

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	v7pushaction.ManifestParser
	ContainsMultipleApps() bool
	InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) error
	ContainsPrivateDockerImages() bool
}

//go:generate counterfeiter . ManifestLocator

type ManifestLocator interface {
	Path(filepathOrDirectory string) (string, bool, error)
}

type PushCommand struct {
	OptionalArgs            flag.OptionalAppName          `positional-args:"yes"`
	HealthCheckTimeout      flag.PositiveInteger          `long:"app-start-timeout" short:"t" description:"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app"`
	Buildpacks              []string                      `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	Disk                    flag.Megabytes                `long:"disk" short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	DockerImage             flag.DockerImage              `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername          string                        `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	DropletPath             flag.PathWithExistenceCheck   `long:"droplet" description:"Path to a tgz file with a pre-staged app"`
	HealthCheckHTTPEndpoint string                        `long:"endpoint"  description:"Valid path on the app for an HTTP health check. Only used when specifying --health-check-type=http"`
	HealthCheckType         flag.HealthCheckType          `long:"health-check-type" short:"u" description:"Application health check type. Defaults to 'port'. 'http' requires a valid endpoint, for example, '/health'."`
	Instances               flag.Instances                `long:"instances" short:"i" description:"Number of instances"`
	PathToManifest          flag.PathWithExistenceCheck   `long:"manifest" short:"f" description:"Path to manifest"`
	Memory                  flag.Megabytes                `long:"memory" short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoManifest              bool                          `long:"no-manifest" description:""`
	NoRoute                 bool                          `long:"no-route" description:"Do not map a route to this app"`
	NoStart                 bool                          `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath                 flag.PathWithExistenceCheck   `long:"path" short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	Stack                   string                        `long:"stack" short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	StartCommand            flag.Command                  `long:"start-command" short:"c" description:"Startup command, set to null to reset to default start command"`
	Strategy                flag.DeploymentStrategy       `long:"strategy" description:"Deployment strategy, either rolling or null."`
	Vars                    []template.VarKV              `long:"var" description:"Variable key value pair for variable substitution, (e.g., name=app1); can specify multiple times"`
	PathsToVarsFiles        []flag.PathWithExistenceCheck `long:"vars-file" description:"Path to a variable substitution file for manifest; can specify multiple times"`
	dockerPassword          interface{}                   `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage                   interface{}                   `usage:"CF_NAME push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND]\n   [-f MANIFEST_PATH | --no-manifest] [--no-start] [-i NUM_INSTANCES]\n   [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT]\n   [-u (process | port | http)]   [--no-route | --random-route]\n   [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]...\n \n  CF_NAME push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME]\n   [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\n   [--no-route | --random-route ] [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]..."`
	envCFStagingTimeout     interface{}                   `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout     interface{}                   `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Config          command.Config
	UI              command.UI
	NOAAClient      v3action.NOAAClient
	Actor           PushActor
	VersionActor    V7ActorForPush
	SharedActor     command.SharedActor
	RouteActor      v7action.RouteActor
	ProgressBar     ProgressBar
	PWD             string
	ManifestLocator ManifestLocator
	ManifestParser  ManifestParser
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

	cmd.ManifestLocator = manifestparser.NewLocator()
	cmd.ManifestParser = manifestparser.NewParser()

	return err
}

func (cmd PushCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	if !cmd.NoManifest {
		if err = cmd.ReadManifest(); err != nil {
			return err
		}
	}

	err = cmd.ValidateFlags()
	if err != nil {
		return err
	}

	flagOverrides, err := cmd.GetFlagOverrides()
	if err != nil {
		return err
	}

	err = cmd.ValidateAllowedFlagsForMultipleApps(cmd.ManifestParser.ContainsMultipleApps())
	if err != nil {
		return err
	}

	flagOverrides.DockerPassword, err = cmd.GetDockerPassword(flagOverrides.DockerUsername, cmd.ManifestParser.ContainsPrivateDockerImages())
	if err != nil {
		return err
	}

	pushPlans, err := cmd.Actor.CreatePushPlans(
		cmd.OptionalArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		cmd.ManifestParser,
		flagOverrides,
	)
	if err != nil {
		return err
	}

	pushPlansStream, eventStream, warningsStream, errorStream := cmd.Actor.PrepareSpace(pushPlans, cmd.ManifestParser)
	appNames, err := cmd.processStreamsFromPrepareSpace(pushPlansStream, eventStream, warningsStream, errorStream)

	if err != nil {
		return err
	}

	if len(appNames) == 0 {
		return translatableerror.AppNameOrManifestRequiredError{}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.announcePushing(appNames, user)

	cmd.UI.DisplayText("Getting app info...")
	log.Info("generating the app plan")

	pushPlans, warnings, err := cmd.Actor.UpdateApplicationSettings(pushPlans)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	log.WithField("number of plans", len(pushPlans)).Debug("completed generating plan")

	for _, plan := range pushPlans {
		log.WithField("app_name", plan.Application.Name).Info("actualizing")
		planStream, eventStream, warningsStream, errorStream := cmd.Actor.Actualize(plan, cmd.ProgressBar)
		err := cmd.processApplyStreams(plan.Application.Name, planStream, eventStream, warningsStream, errorStream)

		if cmd.shouldDisplaySummary(err) {
			summaryErr := cmd.displayAppSummary(plan)
			if summaryErr != nil {
				return summaryErr
			}
		}
		if err != nil {
			return cmd.mapErr(plan.Application.Name, err)
		}
	}

	return nil
}

func (cmd PushCommand) shouldDisplaySummary(err error) bool {
	if err == nil {
		return true
	}
	_, ok := err.(actionerror.AllInstancesCrashedError)
	return ok
}

func (cmd PushCommand) mapErr(appName string, err error) error {
	switch err.(type) {
	case actionerror.AllInstancesCrashedError:
		return translatableerror.ApplicationUnableToStartError{
			AppName:    appName,
			BinaryName: cmd.Config.BinaryName(),
		}
	case actionerror.StartupTimeoutError:
		return translatableerror.StartupTimeoutError{
			AppName:    appName,
			BinaryName: cmd.Config.BinaryName(),
		}
	}
	return err
}

func (cmd PushCommand) announcePushing(appNames []string, user configv3.User) {
	tokens := map[string]interface{}{
		"AppName":   strings.Join(appNames, ", "),
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	}
	singular := "Pushing app {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}..."
	plural := "Pushing apps {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}..."

	if len(appNames) == 1 {
		cmd.UI.DisplayTextWithFlavor(singular, tokens)
	} else {
		cmd.UI.DisplayTextWithFlavor(plural, tokens)
	}
}

func (cmd PushCommand) displayAppSummary(plan v7pushaction.PushPlan) error {
	log.Info("getting application summary info")
	summary, warnings, err := cmd.VersionActor.GetApplicationSummaryByNameAndSpace(
		plan.Application.Name,
		cmd.Config.TargetedSpace().GUID,
		true,
		cmd.RouteActor,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayNewline()
	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	appSummaryDisplayer.AppDisplay(summary, true)
	return nil
}

func (cmd PushCommand) processStreamsFromPrepareSpace(
	pushPlansStream <-chan []v7pushaction.PushPlan,
	eventStream <-chan v7pushaction.Event,
	warningsStream <-chan v7pushaction.Warnings,
	errorStream <-chan error,
) ([]string, error) {
	var namesClosed, eventClosed, warningsClosed, errClosed bool
	var appNames []string
	var err error

	for {
		select {
		case plans, ok := <-pushPlansStream:
			if !ok {
				if !namesClosed {
					log.Debug("processing config stream closed")
				}
				namesClosed = true
				break
			}
			for _, plan := range plans {
				appNames = append(appNames, plan.Application.Name)
			}
		case event, ok := <-eventStream:
			if !ok {
				if !eventClosed {
					log.Debug("processing event stream closed")
				}
				eventClosed = true
				break
			}
			_, err := cmd.processEvent(event, cmd.OptionalArgs.AppName)
			if err != nil {
				return nil, err
			}
		case warnings, ok := <-warningsStream:
			if !ok {
				if !warningsClosed {
					log.Debug("processing warnings stream closed")
				}
				warningsClosed = true
				break
			}
			cmd.UI.DisplayWarnings(warnings)
		case receivedError, ok := <-errorStream:
			if !ok {
				if !errClosed {
					log.Debug("processing error stream closed")
				}
				errClosed = true
				break
			}
			return nil, receivedError
		}

		if namesClosed && eventClosed && warningsClosed && errClosed {
			break
		}
	}

	return appNames, err
}

func (cmd PushCommand) processApplyStreams(
	appName string,
	planStream <-chan v7pushaction.PushPlan,
	eventStream <-chan v7pushaction.Event,
	warningsStream <-chan v7pushaction.Warnings,
	errorStream <-chan error,
) error {
	var planClosed, eventClosed, warningsClosed, errClosed, complete bool

	for {
		select {
		case _, ok := <-planStream:
			if !ok {
				if !planClosed {
					log.Debug("processing config stream closed")
				}
				planClosed = true
				break
			}
		case event, ok := <-eventStream:
			if !ok {
				if !eventClosed {
					log.Debug("processing event stream closed")
				}
				eventClosed = true
				break
			}
			var err error
			complete, err = cmd.processEvent(event, appName)
			if err != nil {
				return err
			}
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
			return err
		}

		if planClosed && eventClosed && warningsClosed && complete {
			break
		}
	}

	return nil
}

func (cmd PushCommand) processEvent(event v7pushaction.Event, appName string) (bool, error) {
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
	case v7pushaction.UploadingApplication:
		cmd.UI.DisplayText("All files found in remote cache; nothing to upload.")
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case v7pushaction.UploadWithArchiveComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.UploadingDroplet:
		cmd.UI.DisplayText("Uploading droplet bits...")
		cmd.ProgressBar.Ready()
	case v7pushaction.UploadDropletComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.StoppingApplication:
		cmd.UI.DisplayText("Stopping Application...")
	case v7pushaction.StoppingApplicationComplete:
		cmd.UI.DisplayText("Application Stopped")
	case v7pushaction.ApplyManifest:
		cmd.UI.DisplayText("Applying manifest...")
	case v7pushaction.ApplyManifestComplete:
		cmd.UI.DisplayText("Manifest applied")
	case v7pushaction.StartingStaging:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Staging app and tracing logs...")
		logStream, errStream, warnings, err := cmd.VersionActor.GetStreamingLogsForApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return false, err
		}
		go cmd.getLogs(logStream, errStream)
	case v7pushaction.StagingComplete:
		cmd.NOAAClient.Close()
	case v7pushaction.RestartingApplication:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor(
			"Waiting for app {{.AppName}} to start...",
			map[string]interface{}{
				"AppName": appName,
			},
		)
	case v7pushaction.StartingDeployment:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor(
			"Starting deployment for app {{.AppName}}...",
			map[string]interface{}{
				"AppName": appName,
			},
		)
	case v7pushaction.WaitingForDeployment:
		cmd.UI.DisplayText("Waiting for app to deploy...")
	case v7pushaction.Complete:
		return true, nil
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false, nil
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

func (cmd PushCommand) ReadManifest() error {
	log.Info("reading manifest if exists")
	pathsToVarsFiles := []string{}
	for _, varfilepath := range cmd.PathsToVarsFiles {
		pathsToVarsFiles = append(pathsToVarsFiles, string(varfilepath))
	}

	readPath := cmd.PWD
	if len(cmd.PathToManifest) != 0 {
		log.WithField("manifestPath", cmd.PathToManifest).Debug("reading '-f' provided manifest")
		readPath = string(cmd.PathToManifest)
	}

	pathToManifest, exists, err := cmd.ManifestLocator.Path(readPath)
	if err != nil {
		return err
	}

	if exists {
		log.WithField("manifestPath", pathToManifest).Debug("path to manifest")
		err = cmd.ManifestParser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, cmd.Vars)
		if err != nil {
			log.Errorln("reading manifest:", err)
			return err
		}

		cmd.UI.DisplayText("Using manifest file {{.Path}}", map[string]interface{}{"Path": pathToManifest})
	}

	return nil
}

func (cmd PushCommand) GetFlagOverrides() (v7pushaction.FlagOverrides, error) {
	return v7pushaction.FlagOverrides{
		Buildpacks:          cmd.Buildpacks,
		Stack:               cmd.Stack,
		Disk:                cmd.Disk.NullUint64,
		DropletPath:         string(cmd.DropletPath),
		DockerImage:         cmd.DockerImage.Path,
		DockerUsername:      cmd.DockerUsername,
		HealthCheckEndpoint: cmd.HealthCheckHTTPEndpoint,
		HealthCheckType:     cmd.HealthCheckType.Type,
		HealthCheckTimeout:  cmd.HealthCheckTimeout.Value, Instances: cmd.Instances.NullInt,
		Memory:            cmd.Memory.NullUint64,
		NoStart:           cmd.NoStart,
		ProvidedAppPath:   string(cmd.AppPath),
		SkipRouteCreation: cmd.NoRoute,
		StartCommand:      cmd.StartCommand.FilteredString,
		Strategy:          cmd.Strategy.Name,
	}, nil
}

func (cmd PushCommand) ValidateAllowedFlagsForMultipleApps(containsMultipleApps bool) error {
	if cmd.OptionalArgs.AppName != "" {
		return nil
	}

	allowedFlagsMultipleApps := !(len(cmd.Buildpacks) > 0 ||
		cmd.Disk.IsSet ||
		cmd.DockerImage.Path != "" ||
		cmd.DockerUsername != "" ||
		cmd.DropletPath != "" ||
		cmd.HealthCheckType.Type != "" ||
		cmd.HealthCheckHTTPEndpoint != "" ||
		cmd.HealthCheckTimeout.Value > 0 ||
		cmd.Instances.IsSet ||
		cmd.Stack != "" ||
		cmd.Memory.IsSet ||
		cmd.AppPath != "" ||
		cmd.NoRoute ||
		cmd.StartCommand.IsSet ||
		cmd.Strategy.Name != "")

	if containsMultipleApps && !allowedFlagsMultipleApps {
		return translatableerror.CommandLineArgsWithMultipleAppsError{}
	}

	return nil
}

func (cmd PushCommand) ValidateFlags() error {
	switch {
	case cmd.DockerUsername != "" && cmd.DockerImage.Path == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--docker-image, -o",
			Arg2: "--docker-username",
		}

	case cmd.DockerImage.Path != "" && cmd.Buildpacks != nil:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--buildpack, -b",
				"--docker-image, -o",
			},
		}

	case cmd.DockerImage.Path != "" && cmd.AppPath != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--docker-image, -o",
				"--path, -p",
			},
		}
	case cmd.DockerImage.Path != "" && cmd.Stack != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--stack, -s",
				"--docker-image, -o",
			},
		}
	case cmd.NoManifest && cmd.PathToManifest != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--manifest, -f",
			},
		}
	case cmd.NoManifest && len(cmd.PathsToVarsFiles) > 0:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--vars-file",
			},
		}
	case cmd.NoManifest && len(cmd.Vars) > 0:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--vars",
			},
		}
	case cmd.HealthCheckType.Type == constant.HTTP && cmd.HealthCheckHTTPEndpoint == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--endpoint",
			Arg2: "--health-check-type=http, -u=http",
		}
	case 0 < len(cmd.HealthCheckHTTPEndpoint) && cmd.HealthCheckType.Type != constant.HTTP:
		return translatableerror.RequiredFlagsError{
			Arg1: "--health-check-type=http, -u=http",
			Arg2: "--endpoint",
		}

	case cmd.DropletPath != "" && (cmd.DockerImage.Path != "" || cmd.DockerUsername != "" || cmd.AppPath != ""):
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--droplet",
				"--docker-image, -o",
				"--docker-username",
				"-p",
			},
		}
	}

	return nil
}
