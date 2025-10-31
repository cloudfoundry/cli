package shared

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/sharedaction"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
)

/*
	AppStager interface extracts the complexity of the asynchronous
    staging process, which is used by several commands (e.g. restage,
    copy-package).
*/

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . AppStager

type AppStager interface {
	StageAndStart(app resources.Application, space configv3.Space, organization configv3.Organization, packageGUID string, opts AppStartOpts) error

	StageApp(app resources.Application, packageGUID string, space configv3.Space) (resources.Droplet, error)

	StartApp(app resources.Application, space configv3.Space, organization configv3.Organization, resourceGuid string, opts AppStartOpts) error
}

type AppStartOpts struct {
	AppAction   constant.ApplicationAction
	MaxInFlight int
	NoWait      bool
	Strategy    constant.DeploymentStrategy
	CanarySteps []resources.CanaryStep
}

type Stager struct {
	Actor    stagingAndStartActor
	UI       command.UI
	Config   command.Config
	LogCache sharedaction.LogCacheClient
}

type stagingAndStartActor interface {
	CreateDeployment(dep resources.Deployment) (string, v7action.Warnings, error)
	GetCurrentUser() (configv3.User, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	PollStart(app resources.Application, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForDeployment(app resources.Application, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
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

func (stager *Stager) StageAndStart(app resources.Application, space configv3.Space, organization configv3.Organization, packageGUID string, opts AppStartOpts) error {
	droplet, err := stager.StageApp(app, packageGUID, space)
	if err != nil {
		return err
	}

	stager.UI.DisplayNewline()

	err = stager.StartApp(app, space, organization, droplet.GUID, opts)
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

func (stager *Stager) StartApp(app resources.Application, space configv3.Space, organization configv3.Organization, resourceGuid string, opts AppStartOpts) error {
	if len(opts.Strategy) > 0 {
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

		dep := resources.Deployment{
			Strategy:      opts.Strategy,
			Relationships: resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: app.GUID}},
		}

		if opts.Strategy == constant.DeploymentStrategyCanary && len(opts.CanarySteps) > 0 {
			dep.Options = resources.DeploymentOpts{CanaryDeploymentOptions: &resources.CanaryDeploymentOptions{Steps: opts.CanarySteps}}
		}
		switch opts.AppAction {
		case constant.ApplicationRollingBack:
			dep.RevisionGUID = resourceGuid
			dep.Options.MaxInFlight = opts.MaxInFlight
			deploymentGUID, warnings, err = stager.Actor.CreateDeployment(dep)
		default:
			dep.DropletGUID = resourceGuid
			dep.Options.MaxInFlight = opts.MaxInFlight
			deploymentGUID, warnings, err = stager.Actor.CreateDeployment(dep)
		}

		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		stager.UI.DisplayText("Waiting for app to deploy...\n")

		handleInstanceDetails := func(instanceDetails string) {
			stager.UI.DisplayText(instanceDetails)
		}

		warnings, err = stager.Actor.PollStartForDeployment(app, deploymentGUID, opts.NoWait, handleInstanceDetails)
		stager.UI.DisplayNewline()
		stager.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		if opts.NoWait && opts.Strategy != constant.DeploymentStrategyCanary {
			stager.UI.DisplayText("First instance restaged correctly, restaging remaining in the background")
			return nil
		}
	} else {
		user, err := stager.Actor.GetCurrentUser()
		if err != nil {
			return err
		}

		flavorText := fmt.Sprintf("%s app {{.App}} in org {{.Org}} / space {{.Space}} as {{.UserName}}...", opts.AppAction)
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
			if opts.AppAction == constant.ApplicationStarting {
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

		warnings, err = stager.Actor.PollStart(app, opts.NoWait, handleInstanceDetails)
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
