package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type RestartActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	StartApplication(appGUID string) (v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	CreateDeployment(appGUID string, dropletGUID string) (string, v7action.Warnings, error)
	PollStart(appGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForRolling(appGUID string, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
}

type RestartCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName            `positional-args:"yes"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy, either rolling or null."`
	NoWait              bool                    `long:"no-wait" description:"Do not wait for the long-running operation to complete; push exits when one instance of the web process is healthy"`
	usage               interface{}             `usage:"CF_NAME restart APP_NAME"`
	relatedCommands     interface{}             `related_commands:"restage, restart-app-instance"`
	envCFStagingTimeout interface{}             `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
}

func (cmd RestartCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Restarting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	switch cmd.Strategy.Name {
	case constant.DeploymentStrategyRolling:
		err = cmd.zeroDowntimeRestart(app)
	default:
		err = cmd.downtimeRestart(app)
	}

	cmd.UI.DisplayNewline()
	if err != nil {
		return err
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

func (cmd RestartCommand) downtimeRestart(app v7action.Application) error {
	var warnings v7action.Warnings
	var err error
	if app.Started() {
		cmd.UI.DisplayText("Stopping app...\n")

		warnings, err = cmd.Actor.StopApplication(app.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	warnings, err = cmd.Actor.StartApplication(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Waiting for app to start...\n")

	handleInstanceDetails := func(instanceDetails string) {
		cmd.UI.DisplayText(instanceDetails)
	}

	warnings, err = cmd.Actor.PollStart(app.GUID, cmd.NoWait, handleInstanceDetails)
	cmd.UI.DisplayWarnings(warnings)
	return err
}

func (cmd RestartCommand) zeroDowntimeRestart(app v7action.Application) error {
	cmd.UI.DisplayText("Creating deployment for app {{.AppName}}...\n",
		map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		},
	)

	deploymentGUID, warnings, err := cmd.Actor.CreateDeployment(app.GUID, "")
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Waiting for app to deploy...\n")

	handleInstanceDetails := func(instanceDetails string) {
		cmd.UI.DisplayText(instanceDetails)
	}

	warnings, err = cmd.Actor.PollStartForRolling(app.GUID, deploymentGUID, cmd.NoWait, handleInstanceDetails)
	cmd.UI.DisplayWarnings(warnings)

	return err
}
