package shared

import (
	"context"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
)

/*
	AppStager interface extracts the complexity of the asynchronous
    staging process, which is used by several commands (e.g. restage,
    copy-package).
*/

//go:generate counterfeiter . AppStager

type AppStager interface {
	StageAndStart(
		app v7action.Application,
		packageGUID string,
		strategy constant.DeploymentStrategy,
		noWait bool,
	) error
}

type Stager struct {
	Actor    stagingAndStartActor
	UI       command.UI
	Config   command.Config
	LogCache sharedaction.LogCacheClient
}

type stagingAndStartActor interface {
	CreateDeployment(appGUID string, dropletGUID string) (string, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	PollStart(appGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForRolling(appGUID string, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
	StagePackage(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error)
	StartApplication(appGUID string) (v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
}

func NewAppStager(actor stagingAndStartActor, ui command.UI, config command.Config, logCache sharedaction.LogCacheClient) AppStager {
	return &Stager{
		Actor:    actor,
		UI:       ui,
		Config:   config,
		LogCache: logCache,
	}
}

func (stager *Stager) StageAndStart(
	app v7action.Application,
	packageGUID string,
	strategy constant.DeploymentStrategy,
	noWait bool,
) error {
	var warnings v7action.Warnings

	logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := stager.Actor.GetStreamingLogsForApplicationByNameAndSpace(app.Name, stager.Config.TargetedSpace().GUID, stager.LogCache)
	stager.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return logErr
	}
	defer stopLogStreamFunc()

	stager.UI.DisplayText("Staging app and tracing logs...")
	dropletStream, warningsStream, errStream := stager.Actor.StagePackage(
		packageGUID,
		app.Name,
		stager.Config.TargetedSpace().GUID,
	)

	droplet, err := PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, stager.UI)
	if err != nil {
		return err
	}

	if strategy == constant.DeploymentStrategyRolling {
		stager.UI.DisplayText("Creating deployment for app {{.AppName}}...\n",
			map[string]interface{}{
				"AppName": app.Name,
			},
		)
		deploymentGUID, warnings, err := stager.Actor.CreateDeployment(app.GUID, droplet.GUID)
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		stager.UI.DisplayText("Waiting for app to deploy...\n")

		handleInstanceDetails := func(instanceDetails string) {
			stager.UI.DisplayText(instanceDetails)
		}

		warnings, err = stager.Actor.PollStartForRolling(app.GUID, deploymentGUID, noWait, handleInstanceDetails)
		stager.UI.DisplayNewline()
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	} else {
		warnings, err = stager.Actor.StopApplication(app.GUID)
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		// attach droplet to app
		warnings, err = stager.Actor.SetApplicationDroplet(app.GUID, droplet.GUID)
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		stager.UI.DisplayNewline()
		stager.UI.DisplayText("Waiting for app to start...")
		stager.UI.DisplayNewline()

		// start the application
		warnings, err = stager.Actor.StartApplication(app.GUID)
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		handleInstanceDetails := func(instanceDetails string) {
			stager.UI.DisplayText(instanceDetails)
		}

		warnings, err = stager.Actor.PollStart(app.GUID, noWait, handleInstanceDetails)
		stager.UI.DisplayNewline()
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	summary, warnings, err := stager.Actor.GetDetailedAppSummary(
		app.Name,
		stager.Config.TargetedSpace().GUID,
		false,
	)
	stager.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appSummaryDisplayer := NewAppSummaryDisplayer(stager.UI)
	appSummaryDisplayer.AppDisplay(summary, false)

	return nil
}
