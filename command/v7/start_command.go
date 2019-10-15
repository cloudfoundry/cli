package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	v6shared "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . StartActor

type StartActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	PollStart(appGUID string, noWait bool) (v7action.Warnings, error)
	StartApplication(appGUID string) (v7action.Warnings, error)
	GetUnstagedNewestPackageGUID(appGuid string) (string, v7action.Warnings, error)
	StagePackage(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
}

type StartCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME"`
	relatedCommands     interface{}  `related_commands:"apps, logs, scale, ssh, stop, restart, run-task"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI          command.UI
	Config      command.Config
	NOAAClient  v7action.NOAAClient
	SharedActor command.SharedActor
	Actor       StartActor
}

func (cmd *StartCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	cmd.NOAAClient = v6shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

	return nil
}

func (cmd StartCommand) Execute(args []string) error {
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

	if app.Started() {
		cmd.UI.DisplayText("App '{{.AppName}}' is already started.",
			map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
		cmd.UI.DisplayOK()
		return nil
	}

	var droplet v7action.Droplet
	packageGuid, warnings, err := cmd.Actor.GetUnstagedNewestPackageGUID(app.GUID)
	if packageGuid != "" {
		dropletStream, warningsStream, errStream := cmd.Actor.StagePackage(packageGuid, cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)

		logStream, logErrStream, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
		cmd.UI.DisplayWarnings(logWarnings)
		if logErr != nil {
			return logErr
		}

		droplet, err = shared.PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Package staged")
	}

	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	warnings, err = cmd.Actor.StartApplication(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayText("\nWaiting for app to start...")

	warnings, err = cmd.Actor.PollStart(app.GUID, false)
	cmd.UI.DisplayWarnings(warnings)
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
