package v7

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v6shared "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/progressbar"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

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
	HandleFlagOverrides(baseManifest pushmanifestparser.Manifest, flagOverrides v7pushaction.FlagOverrides) (pushmanifestparser.Manifest, error)
	CreatePushPlans(spaceGUID string, orgGUID string, manifest pushmanifestparser.Manifest, overrides v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, v7action.Warnings, error)
	// Actualize applies any necessary changes.
	Actualize(plan v7pushaction.PushPlan, progressBar v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent
}

//go:generate counterfeiter . V7ActorForPush

type V7ActorForPush interface {
	AppActor
	SetSpaceManifest(spaceGUID string, rawManifest []byte) (v7action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	RestartApplication(appGUID string, noWait bool) (v7action.Warnings, error)
}

//go:generate counterfeiter . PushManifestParser

type PushManifestParser interface {
	InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (pushmanifestparser.Manifest, error)
	MarshalManifest(manifest pushmanifestparser.Manifest) ([]byte, error)
}

//go:generate counterfeiter . ManifestLocator

type ManifestLocator interface {
	Path(filepathOrDirectory string) (string, bool, error)
}

type PushCommand struct {
	OptionalArgs            flag.OptionalAppName                `positional-args:"yes"`
	HealthCheckTimeout      flag.PositiveInteger                `long:"app-start-timeout" short:"t" description:"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app"`
	Buildpacks              []string                            `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	Disk                    string                              `long:"disk" short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	DockerImage             flag.DockerImage                    `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername          string                              `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	DropletPath             flag.PathWithExistenceCheck         `long:"droplet" description:"Path to a tgz file with a pre-staged app"`
	HealthCheckHTTPEndpoint string                              `long:"endpoint"  description:"Valid path on the app for an HTTP health check. Only used when specifying --health-check-type=http"`
	HealthCheckType         flag.HealthCheckType                `long:"health-check-type" short:"u" description:"Application health check type. Defaults to 'port'. 'http' requires a valid endpoint, for example, '/health'."`
	Instances               flag.Instances                      `long:"instances" short:"i" description:"Number of instances"`
	PathToManifest          flag.ManifestPathWithExistenceCheck `long:"manifest" short:"f" description:"Path to manifest"`
	Memory                  string                              `long:"memory" short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoManifest              bool                                `long:"no-manifest" description:"Ignore manifest file"`
	NoRoute                 bool                                `long:"no-route" description:"Do not map a route to this app"`
	NoStart                 bool                                `long:"no-start" description:"Do not stage and start the app after pushing"`
	NoWait                  bool                                `long:"no-wait" description:"Do not wait for the long-running operation to complete; push exits when one instance of the web process is healthy"`
	AppPath                 flag.PathWithExistenceCheck         `long:"path" short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	RandomRoute             bool                                `long:"random-route" description:"Create a random route for this app (except when no-route is specified in the manifest)"`
	Stack                   string                              `long:"stack" short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	StartCommand            flag.Command                        `long:"start-command" short:"c" description:"Startup command, set to null to reset to default start command"`
	Strategy                flag.DeploymentStrategy             `long:"strategy" description:"Deployment strategy, either rolling or null."`
	Vars                    []template.VarKV                    `long:"var" description:"Variable key value pair for variable substitution, (e.g., name=app1); can specify multiple times"`
	PathsToVarsFiles        []flag.PathWithExistenceCheck       `long:"vars-file" description:"Path to a variable substitution file for manifest; can specify multiple times"`
	dockerPassword          interface{}                         `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage                   interface{}                         `usage:"CF_NAME push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND]\n   [-f MANIFEST_PATH | --no-manifest] [--no-start] [--no-wait] [-i NUM_INSTANCES]\n   [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT]\n   [-u (process | port | http)]   [--no-route | --random-route]\n   [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]...\n \n  CF_NAME push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME]\n   [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start] [--no-wait]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\n   [--no-route | --random-route ] [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]..."`
	envCFStagingTimeout     interface{}                         `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout     interface{}                         `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Config          command.Config
	UI              command.UI
	NOAAClient      v3action.NOAAClient
	Actor           PushActor
	VersionActor    V7ActorForPush
	SharedActor     command.SharedActor
	ProgressBar     ProgressBar
	PWD             string
	ManifestLocator ManifestLocator
	ManifestParser  PushManifestParser
}

func (cmd *PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.ProgressBar = progressbar.NewProgressBar()

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}

	v7actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	cmd.VersionActor = v7actor
	cmd.Actor = v7pushaction.NewActor(v7actor, sharedActor)

	cmd.NOAAClient = v6shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

	currentDir, err := os.Getwd()
	cmd.PWD = currentDir

	cmd.ManifestLocator = pushmanifestparser.NewLocator()
	cmd.ManifestParser = pushmanifestparser.ManifestParser{}

	return err
}

func (cmd PushCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	flagOverrides, err := cmd.GetFlagOverrides()
	if err != nil {
		return err
	}

	err = cmd.ValidateFlags()
	if err != nil {
		return err
	}

	baseManifest, err := cmd.GetBaseManifest(flagOverrides)
	if err != nil {
		return err
	}

	transformedManifest, err := cmd.Actor.HandleFlagOverrides(baseManifest, flagOverrides)
	if err != nil {
		return err
	}

	flagOverrides.DockerPassword, err = cmd.GetDockerPassword(flagOverrides.DockerUsername, transformedManifest.ContainsPrivateDockerImages())
	if err != nil {
		return err
	}

	transformedRawManifest, err := cmd.ManifestParser.MarshalManifest(transformedManifest)
	if err != nil {
		return err
	}

	cmd.announcePushing(transformedManifest.AppNames(), user)

	hasManifest := transformedManifest.PathToManifest != ""

	if hasManifest {
		cmd.UI.DisplayText("Applying manifest file {{.Path}}...", map[string]interface{}{
			"Path": transformedManifest.PathToManifest,
		})
	}

	v7ActionWarnings, err := cmd.VersionActor.SetSpaceManifest(
		cmd.Config.TargetedSpace().GUID,
		transformedRawManifest,
	)

	cmd.UI.DisplayWarnings(v7ActionWarnings)
	if err != nil {
		return err
	}
	if hasManifest {
		cmd.UI.DisplayText("Manifest applied")
	}

	pushPlans, warnings, err := cmd.Actor.CreatePushPlans(
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		transformedManifest,
		flagOverrides,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	log.WithField("number of plans", len(pushPlans)).Debug("completed generating plan")

	for _, plan := range pushPlans {
		log.WithField("app_name", plan.Application.Name).Info("actualizing")
		eventStream := cmd.Actor.Actualize(plan, cmd.ProgressBar)
		err := cmd.eventStreamHandler(eventStream)

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

func (cmd PushCommand) GetBaseManifest(flagOverrides v7pushaction.FlagOverrides) (pushmanifestparser.Manifest, error) {
	defaultManifest := pushmanifestparser.Manifest{
		Applications: []pushmanifestparser.Application{
			{Name: flagOverrides.AppName},
		},
	}
	if cmd.NoManifest {
		return defaultManifest, nil
	}

	log.Info("reading manifest if exists")
	var pathsToVarsFiles []string
	pathsToVarsFiles = append(pathsToVarsFiles, flagOverrides.PathsToVarsFiles...)

	readPath := cmd.PWD
	if flagOverrides.ManifestPath != "" {
		log.WithField("manifestPath", flagOverrides.ManifestPath).Debug("reading '-f' provided manifest")
		readPath = flagOverrides.ManifestPath
	}

	pathToManifest, exists, err := cmd.ManifestLocator.Path(readPath)
	if err != nil {
		return pushmanifestparser.Manifest{}, err
	}

	if !exists {
		return defaultManifest, nil
	}

	log.WithField("manifestPath", pathToManifest).Debug("path to manifest")
	manifest, err := cmd.ManifestParser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, flagOverrides.Vars)
	if err != nil {
		log.Errorln("reading manifest:", err)
		return pushmanifestparser.Manifest{}, err
	}

	return manifest, nil
}

func (cmd PushCommand) GetDockerPassword(dockerUsername string, containsPrivateDockerImages bool) (string, error) {
	if dockerUsername == "" && !containsPrivateDockerImages { // no need for a password without a username
		return "", nil
	}

	if cmd.Config.DockerPassword() == "" {
		cmd.UI.DisplayText("Environment variable CF_DOCKER_PASSWORD not set.")
		return cmd.UI.DisplayPasswordPrompt("Docker password")
	}

	cmd.UI.DisplayText("Using docker repository password from environment variable CF_DOCKER_PASSWORD.")
	return cmd.Config.DockerPassword(), nil
}

func (cmd PushCommand) GetFlagOverrides() (v7pushaction.FlagOverrides, error) {
	var pathsToVarsFiles []string
	for _, varFilePath := range cmd.PathsToVarsFiles {
		pathsToVarsFiles = append(pathsToVarsFiles, string(varFilePath))
	}

	return v7pushaction.FlagOverrides{
		AppName:             cmd.OptionalArgs.AppName,
		Buildpacks:          cmd.Buildpacks,
		Stack:               cmd.Stack,
		Disk:                cmd.Disk,
		DropletPath:         string(cmd.DropletPath),
		DockerImage:         cmd.DockerImage.Path,
		DockerUsername:      cmd.DockerUsername,
		HealthCheckEndpoint: cmd.HealthCheckHTTPEndpoint,
		HealthCheckType:     cmd.HealthCheckType.Type,
		HealthCheckTimeout:  cmd.HealthCheckTimeout.Value,
		Instances:           cmd.Instances.NullInt,
		Memory:              cmd.Memory,
		NoStart:             cmd.NoStart,
		NoWait:              cmd.NoWait,
		ProvidedAppPath:     string(cmd.AppPath),
		NoRoute:             cmd.NoRoute,
		RandomRoute:         cmd.RandomRoute,
		StartCommand:        cmd.StartCommand.FilteredString,
		Strategy:            cmd.Strategy.Name,
		ManifestPath:        string(cmd.PathToManifest),
		PathsToVarsFiles:    pathsToVarsFiles,
		Vars:                cmd.Vars,
		NoManifest:          cmd.NoManifest,
	}, nil
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

	case cmd.NoStart && cmd.Strategy == flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-start",
				"--strategy=rolling",
			},
		}

	case cmd.NoStart && cmd.NoWait:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-start",
				"--no-wait",
			},
		}

	case cmd.NoRoute && cmd.RandomRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-route",
				"--random-route",
			},
		}
	case !cmd.validBuildpacks():
		return translatableerror.InvalidBuildpacksError{}
	}

	return nil
}

func (cmd PushCommand) validBuildpacks() bool {
	for _, buildpack := range cmd.Buildpacks {
		if (buildpack == "null" || buildpack == "default") && len(cmd.Buildpacks) > 1 {
			return false
		}
	}
	return true
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
	summary, warnings, err := cmd.VersionActor.GetDetailedAppSummary(
		plan.Application.Name,
		cmd.Config.TargetedSpace().GUID,
		true,
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

func (cmd PushCommand) eventStreamHandler(eventStream <-chan *v7pushaction.PushEvent) error {
	for event := range eventStream {
		cmd.UI.DisplayWarnings(event.Warnings)
		if event.Err != nil {
			return event.Err
		}
		err := cmd.processEvent(event.Event, event.Plan.Application.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd PushCommand) processEvent(event v7pushaction.Event, appName string) error {
	switch event {
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
			return err
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
	default:
		log.WithField("event", event).Debug("ignoring event")
	}

	return nil
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
