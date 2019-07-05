package v6

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . V3ZeroDowntimeVersionActor

type V3ZeroDowntimeVersionActor interface {
	ZeroDowntimePollStart(appGUID string, warningsChannel chan<- v3action.Warnings) error
	CreateDeployment(appGUID string, deploymentGUID string) (string, v3action.Warnings, error)
	PollDeployment(deploymentGUID string, warningsChannel chan<- v3action.Warnings) error
	CloudControllerAPIVersion() string
	CreateAndUploadBitsPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (v3action.Package, v3action.Warnings, error)
	CreateDockerPackageByApplicationNameAndSpace(appName string, spaceGUID string, dockerImageCredentials v3action.DockerImageCredentials) (v3action.Package, v3action.Warnings, error)
	CreateApplicationInSpace(app v3action.Application, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetCurrentDropletByApplication(appGUID string) (v3action.Droplet, v3action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error)
	PollStart(appGUID string, warningsChannel chan<- v3action.Warnings) error
	SetApplicationDropletByApplicationNameAndSpace(appName string, spaceGUID string, dropletGUID string) (v3action.Warnings, error)
	StagePackage(packageGUID string, appName string) (<-chan v3action.Droplet, <-chan v3action.Warnings, <-chan error)
	RestartApplication(appGUID string) (v3action.Warnings, error)
	UpdateApplication(app v3action.Application) (v3action.Application, v3action.Warnings, error)
}

type V3ZeroDowntimePushCommand struct {
	RequiredArgs        flag.AppName                `positional-args:"yes"`
	Buildpacks          []string                    `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	StackName           string                      `short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	DockerImage         flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername      string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	NoRoute             bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart             bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	WaitUntilDeployed   bool                        `long:"wait-for-deploy-complete" description:"Wait for the entire deployment to complete"`
	AppPath             flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	dockerPassword      interface{}                 `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage               interface{}                 `usage:"CF_NAME v3-zdt-push APP_NAME [-b BUILDPACK]... [-p APP_PATH] [--no-route] [--no-start]\n   CF_NAME v3-zdt-push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME] [--no-route] [--no-start]"`
	envCFStagingTimeout interface{}                 `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}                 `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                  command.UI
	Config              command.Config
	NOAAClient          v3action.NOAAClient
	SharedActor         command.SharedActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
	PackageDisplayer    shared.PackageDisplayer
	ProgressBar         ProgressBar

	ZdtActor            V3ZeroDowntimeVersionActor
	V3PushActor         V3PushActor
	OriginalV2PushActor OriginalV2PushActor
}

func (cmd *V3ZeroDowntimePushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		return err
	}
	v3actor := v3action.NewActor(ccClient, config, sharedActor, nil)
	cmd.ZdtActor = v3actor
	cmd.V3PushActor = v3actor

	ccClientV2, uaaClientV2, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClientV2, uaaClientV2, config)

	cmd.SharedActor = sharedActor
	cmd.OriginalV2PushActor = pushaction.NewActor(v2Actor, v3actor, sharedActor)

	v2AppActor := v2action.NewActor(ccClientV2, uaaClientV2, config)
	cmd.NOAAClient = shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

	cmd.AppSummaryDisplayer = shared.AppSummaryDisplayer{
		UI:         cmd.UI,
		Config:     cmd.Config,
		Actor:      cmd.V3PushActor,
		V2AppActor: v2AppActor,
		AppName:    cmd.RequiredArgs.AppName,
	}
	cmd.PackageDisplayer = shared.NewPackageDisplayer(cmd.UI, cmd.Config)

	return nil
}

func (cmd V3ZeroDowntimePushCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.validateArgs()
	if err != nil {
		return err
	}

	err = command.MinimumCCAPIVersionCheck(cmd.ZdtActor.CloudControllerAPIVersion(), ccversion.MinVersionZeroDowntimePushV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !verifyBuildpacks(cmd.Buildpacks) {
		return translatableerror.ConflictingBuildpacksError{}
	}

	var app v3action.Application
	app, err = cmd.getApplication()
	if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
		app, err = cmd.createApplication(user.Name)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		app, err = cmd.updateApplication(user.Name, app.GUID)
		if err != nil {
			return err
		}
	}

	pkg, err := cmd.createPackage()
	if err != nil {
		return err
	}

	if cmd.NoStart {
		return nil
	}

	dropletGUID, err := cmd.stagePackage(pkg, user.Name)
	if err != nil {
		return err
	}

	if !cmd.NoRoute {
		err = cmd.createAndMapRoutes(app)
		if err != nil {
			return err
		}
	}

	warnings := make(chan v3action.Warnings)
	done := make(chan bool)
	go func() {
		for {
			select {
			case message := <-warnings:
				cmd.UI.DisplayWarnings(message)
			case <-done:
				return
			}
		}
	}()

	switch app.State {
	case constant.ApplicationStopped:
		err = cmd.setApplicationDroplet(dropletGUID, user.Name)
		if err != nil {
			return err
		}

		err = cmd.restartApplication(app.GUID, user.Name)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("Waiting for app to start...")
		err = cmd.ZdtActor.PollStart(app.GUID, warnings)

	case constant.ApplicationStarted:
		var deploymentGUID string
		deploymentGUID, err = cmd.createDeployment(app.GUID, user.Name, dropletGUID)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("Waiting for app to start...")
		if cmd.WaitUntilDeployed {
			err = cmd.ZdtActor.PollDeployment(deploymentGUID, warnings) //
		} else {
			err = cmd.ZdtActor.ZeroDowntimePollStart(app.GUID, warnings)
		}
	default:
		return fmt.Errorf("inconceivable application state: %s", app.State)
	}

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

	cmd.UI.DisplayTextWithFlavor("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	return cmd.AppSummaryDisplayer.DisplayAppInfo()
}

func (cmd V3ZeroDowntimePushCommand) validateArgs() error {
	switch {
	case cmd.DockerImage.Path != "" && cmd.AppPath != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--docker-image", "-o", "-p"},
		}
	case cmd.DockerImage.Path != "" && len(cmd.Buildpacks) > 0:
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-b", "--docker-image", "-o"},
		}
	case cmd.DockerUsername != "" && cmd.DockerImage.Path == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--docker-image, -o", Arg2: "--docker-username",
		}
	case cmd.DockerUsername != "" && cmd.Config.DockerPassword() == "":
		return translatableerror.DockerPasswordNotSetError{}
	}
	return nil
}

func (cmd V3ZeroDowntimePushCommand) createApplication(userName string) (v3action.Application, error) {
	appToCreate := v3action.Application{
		Name: cmd.RequiredArgs.AppName,
	}

	if cmd.DockerImage.Path != "" {
		appToCreate.LifecycleType = constant.AppLifecycleTypeDocker
	} else {
		appToCreate.LifecycleType = constant.AppLifecycleTypeBuildpack
		appToCreate.LifecycleBuildpacks = cmd.Buildpacks
		appToCreate.StackName = cmd.StackName
	}

	app, warnings, err := cmd.ZdtActor.CreateApplicationInSpace(
		appToCreate,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v3action.Application{}, err
	}

	cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  userName,
	})

	cmd.UI.DisplayOK()
	return app, nil
}

func (cmd V3ZeroDowntimePushCommand) getApplication() (v3action.Application, error) {
	app, warnings, err := cmd.ZdtActor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v3action.Application{}, err
	}

	return app, nil
}

func (cmd V3ZeroDowntimePushCommand) updateApplication(userName string, appGUID string) (v3action.Application, error) {
	cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  userName,
	})

	appToUpdate := v3action.Application{
		GUID: appGUID,
	}

	if cmd.DockerImage.Path != "" {
		appToUpdate.LifecycleType = constant.AppLifecycleTypeDocker

	} else {
		appToUpdate.LifecycleType = constant.AppLifecycleTypeBuildpack
		appToUpdate.LifecycleBuildpacks = cmd.Buildpacks
		appToUpdate.StackName = cmd.StackName
	}

	app, warnings, err := cmd.ZdtActor.UpdateApplication(appToUpdate)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v3action.Application{}, err
	}

	cmd.UI.DisplayOK()
	return app, nil
}

func (cmd V3ZeroDowntimePushCommand) createAndMapRoutes(app v3action.Application) error {
	cmd.UI.DisplayText("Mapping routes...")
	routeWarnings, err := cmd.OriginalV2PushActor.CreateAndMapDefaultApplicationRoute(cmd.Config.TargetedOrganization().GUID, cmd.Config.TargetedSpace().GUID, v2action.Application{Name: app.Name, GUID: app.GUID})
	cmd.UI.DisplayWarnings(routeWarnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd V3ZeroDowntimePushCommand) createPackage() (v3action.Package, error) {
	isDockerImage := cmd.DockerImage.Path != ""
	err := cmd.PackageDisplayer.DisplaySetupMessage(cmd.RequiredArgs.AppName, isDockerImage)
	if err != nil {
		return v3action.Package{}, err
	}

	var (
		pkg      v3action.Package
		warnings v3action.Warnings
	)

	if isDockerImage {
		pkg, warnings, err = cmd.ZdtActor.CreateDockerPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, v3action.DockerImageCredentials{Path: cmd.DockerImage.Path, Username: cmd.DockerUsername, Password: cmd.Config.DockerPassword()})
	} else {
		pkg, warnings, err = cmd.ZdtActor.CreateAndUploadBitsPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, string(cmd.AppPath))
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v3action.Package{}, err
	}

	cmd.UI.DisplayOK()
	return pkg, nil
}

func (cmd V3ZeroDowntimePushCommand) stagePackage(pkg v3action.Package, userName string) (string, error) {
	cmd.UI.DisplayTextWithFlavor("Staging package for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  userName,
	})

	logStream, logErrStream, logWarnings, logErr := cmd.ZdtActor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
	cmd.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return "", logErr
	}

	buildStream, warningsStream, errStream := cmd.ZdtActor.StagePackage(pkg.GUID, cmd.RequiredArgs.AppName)
	droplet, err := shared.PollStage(buildStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
	if err != nil {
		return "", err
	}

	cmd.UI.DisplayOK()
	return droplet.GUID, nil
}

func (cmd V3ZeroDowntimePushCommand) setApplicationDroplet(dropletGUID string, userName string) error {
	cmd.UI.DisplayTextWithFlavor("Setting app {{.AppName}} to droplet {{.DropletGUID}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"DropletGUID": dropletGUID,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   cmd.Config.TargetedSpace().Name,
		"Username":    userName,
	})

	warnings, err := cmd.ZdtActor.SetApplicationDropletByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, dropletGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd V3ZeroDowntimePushCommand) restartApplication(appGUID string, userName string) error {
	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  userName,
	})

	warnings, err := cmd.ZdtActor.RestartApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	return nil
}

func (cmd V3ZeroDowntimePushCommand) createDeployment(appGUID string, userName string, dropletGUID string) (string, error) {
	cmd.UI.DisplayTextWithFlavor("Starting deployment for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  userName,
	})

	deploymentGUID, warnings, err := cmd.ZdtActor.CreateDeployment(appGUID, dropletGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return "", err
	}
	cmd.UI.DisplayOK()
	return deploymentGUID, nil
}
