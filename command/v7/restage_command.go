package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
	"context"
)

//go:generate counterfeiter . RestageActor

type RestageActor interface {
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.LogCacheClient) (<-chan v7action.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetNewestReadyPackageForApplication(appGUID string) (v7action.Package, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
	StagePackage(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error)
	StartApplication(appGUID string) (v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	PollStart(appGUID string, noWait bool) (v7action.Warnings, error)
	CreateDeployment(appGUID string, dropletGUID string) (string, v7action.Warnings, error)
	PollStartForRolling(appGUID string, deploymentGUID string, noWait bool) (v7action.Warnings, error)
}

type RestageCommand struct {
	RequiredArgs        flag.AppName            `positional-args:"yes"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy, either rolling or null."`
	NoWait              bool                    `long:"no-wait" description:"Do not wait for the long-running operation to complete; push exits when one instance of the web process is healthy"`
	usage               interface{}             `usage:"CF_NAME restage APP_NAME\n\nEXAMPLES:\n   CF_NAME restage APP_NAME\n   CF_NAME restage APP_NAME --strategy=rolling\n   CF_NAME restage APP_NAME --strategy=rolling --no-wait\n"`
	relatedCommands     interface{}             `related_commands:"restart"`
	envCFStagingTimeout interface{}             `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI             command.UI
	Config         command.Config
	SharedActor    command.SharedActor
	LogCacheClient v7action.LogCacheClient
	Actor          RestageActor
}

func (cmd *RestageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	cmd.LogCacheClient = shared.NewLogCacheClient(ccClient.Info.LogCache(), config, ui)

	return nil
}

func (cmd RestageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if cmd.Strategy.Name != constant.DeploymentStrategyRolling {
		cmd.UI.DisplayWarning("This action will cause app downtime.")
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTextWithFlavor("Restaging app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	pkg, warnings, err := cmd.Actor.GetNewestReadyPackageForApplication(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return cmd.mapErr(cmd.RequiredArgs.AppName, err)
	}

	logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.LogCacheClient)
	cmd.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return logErr
	}
	defer stopLogStreamFunc()

	cmd.UI.DisplayText("Staging app and tracing logs...")
	dropletStream, warningsStream, errStream := cmd.Actor.StagePackage(
		pkg.GUID,
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	droplet, err := shared.PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
	if err != nil {
		return cmd.mapErr(cmd.RequiredArgs.AppName, err)
	}

	if cmd.Strategy.Name == constant.DeploymentStrategyRolling {
		cmd.UI.DisplayText("Creating deployment for app {{.AppName}}...\n",
			map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			},
		)
		deploymentGUID, warnings, err := cmd.Actor.CreateDeployment(app.GUID, droplet.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("Waiting for app to deploy...\n")
		warnings, err = cmd.Actor.PollStartForRolling(app.GUID, deploymentGUID, cmd.NoWait)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return cmd.mapErr(cmd.RequiredArgs.AppName, err)
		}
	} else {
		warnings, err = cmd.Actor.StopApplication(app.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		// attach droplet to app
		warnings, err = cmd.Actor.SetApplicationDroplet(app.GUID, droplet.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for app to start...")
		cmd.UI.DisplayNewline()

		// start the application
		warnings, err = cmd.Actor.StartApplication(app.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		warnings, err = cmd.Actor.PollStart(app.GUID, cmd.NoWait)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return cmd.mapErr(cmd.RequiredArgs.AppName, err)
		}
	}

	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	summary, warnings, err := cmd.Actor.GetDetailedAppSummary(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		false,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	appSummaryDisplayer.AppDisplay(summary, false)

	return nil
}

func (cmd RestageCommand) mapErr(appName string, err error) error {
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
	case actionerror.StagingFailedNoAppDetectedError:
		return translatableerror.StagingFailedNoAppDetectedError{
			Message:    err.Error(),
			BinaryName: cmd.Config.BinaryName(),
		}
	case actionerror.PackageNotFoundInAppError:
		return translatableerror.PackageNotFoundInAppError{
			AppName:    cmd.RequiredArgs.AppName,
			BinaryName: cmd.Config.BinaryName(),
		}
	}
	return err
}
