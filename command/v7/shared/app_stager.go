package shared

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
)

/*
	AppStager interface extracts the complexity of the asynchronous
    staging process, which is used by several commands (e.g. restage,
    copy-package).
*/

//go:generate counterfeiter . AppStager

type AppStager interface {
	StageAndStart(
		app resources.Application,
		space configv3.Space,
		organization configv3.Organization,
		packageGUID string,
		strategy constant.DeploymentStrategy,
		noWait bool,
		appAction constant.ApplicationAction,
	) error

	StageApp(
		app resources.Application,
		packageGUID string,
		space configv3.Space,
	) (resources.Droplet, error)

	StartApp(
		app resources.Application,
		resourceGuid string,
		strategy constant.DeploymentStrategy,
		noWait bool,
		space configv3.Space,
		organization configv3.Organization,
		appAction constant.ApplicationAction,
	) error
}

type Stager struct {
	Actor    stagingAndStartActor
	UI       command.UI
	Config   command.Config
	LogCache sharedaction.LogCacheClient
}

type stagingAndStartActor interface {
	CreateDeploymentByApplicationAndDroplet(appGUID string, dropletGUID string) (string, v7action.Warnings, error)
	CreateDeploymentByApplicationAndRevision(appGUID string, revisionGUID string) (string, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	PollStart(app resources.Application, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForRolling(app resources.Application, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
	StagePackage(packageGUID, appName, spaceGUID string) (<-chan resources.Droplet, <-chan v7action.Warnings, <-chan error)
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
	app resources.Application,
	space configv3.Space,
	organization configv3.Organization,
	packageGUID string,
	strategy constant.DeploymentStrategy,
	noWait bool,
	appAction constant.ApplicationAction,
) error {

	droplet, err := stager.StageApp(app, packageGUID, space)
	if err != nil {
		return err
	}

	stager.UI.DisplayNewline()

	err = stager.StartApp(app, droplet.GUID, strategy, noWait, space, organization, appAction)
	if err != nil {
		return err
	}

	return nil
}

func (stager *Stager) StageApp(app resources.Application, packageGUID string, space configv3.Space) (resources.Droplet, error) {
	logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := stager.Actor.GetStreamingLogsForApplicationByNameAndSpace(app.Name, space.GUID, stager.LogCache)
	stager.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return resources.Droplet{}, logErr
	}
	defer stopLogStreamFunc()

	stager.UI.DisplayText("Staging app and tracing logs...")
	dropletStream, warningsStream, errStream := stager.Actor.StagePackage(
		packageGUID,
		app.Name,
		space.GUID,
	)

	droplet, err := PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, stager.UI)
	if err != nil {
		return resources.Droplet{}, err
	}

	return droplet, nil
}

func (stager *Stager) StartApp(
	app resources.Application,
	resourceGuid string,
	strategy constant.DeploymentStrategy,
	noWait bool,
	space configv3.Space,
	organization configv3.Organization,
	appAction constant.ApplicationAction,
) error {
	if strategy == constant.DeploymentStrategyRolling {
		stager.UI.DisplayText("Creating deployment for app {{.AppName}}...\n",
			map[string]interface{}{
				"AppName": app.Name,
			},
		)

		var (
			deploymentGUID string
			warnings       v7action.Warnings
			err            error
		)

		switch appAction {
		case constant.ApplicationRollingBack:
			deploymentGUID, warnings, err = stager.Actor.CreateDeploymentByApplicationAndRevision(app.GUID, resourceGuid)
		default:
			deploymentGUID, warnings, err = stager.Actor.CreateDeploymentByApplicationAndDroplet(app.GUID, resourceGuid)
		}

		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		stager.UI.DisplayText("Waiting for app to deploy...\n")

		handleInstanceDetails := func(instanceDetails string) {
			stager.UI.DisplayText(instanceDetails)
		}

		warnings, err = stager.Actor.PollStartForRolling(app, deploymentGUID, noWait, handleInstanceDetails)
		stager.UI.DisplayNewline()
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	} else {
		user, err := stager.Config.CurrentUser()
		if err != nil {
			return err
		}

		flavorText := fmt.Sprintf("%s app {{.App}} in org {{.Org}} / space {{.Space}} as {{.UserName}}...", appAction)
		stager.UI.DisplayTextWithFlavor(flavorText,
			map[string]interface{}{
				"App":      app.Name,
				"Org":      organization.Name,
				"Space":    space.Name,
				"UserName": user.Name,
			},
		)
		stager.UI.DisplayNewline()

		if app.Started() {
			if appAction == constant.ApplicationStarting {
				stager.UI.DisplayText("App '{{.AppName}}' is already started.",
					map[string]interface{}{
						"AppName": app.Name,
					})
				stager.UI.DisplayOK()
				return nil

			} else {
				stager.UI.DisplayText("Stopping app...")
				stager.UI.DisplayNewline()

				warnings, err := stager.Actor.StopApplication(app.GUID)
				stager.UI.DisplayWarnings(warnings)
				if err != nil {
					return err
				}
			}
		}

		if resourceGuid != "" {
			// attach droplet to app
			warnings, err := stager.Actor.SetApplicationDroplet(app.GUID, resourceGuid)
			stager.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}
		}

		stager.UI.DisplayText("Waiting for app to start...")
		stager.UI.DisplayNewline()

		// start the application
		warnings, err := stager.Actor.StartApplication(app.GUID)
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		handleInstanceDetails := func(instanceDetails string) {
			stager.UI.DisplayText(instanceDetails)
		}

		warnings, err = stager.Actor.PollStart(app, noWait, handleInstanceDetails)
		stager.UI.DisplayNewline()
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	summary, warnings, err := stager.Actor.GetDetailedAppSummary(
		app.Name,
		space.GUID,
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
