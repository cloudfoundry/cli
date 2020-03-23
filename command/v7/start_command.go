package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type StartCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME"`
	relatedCommands     interface{}  `related_commands:"apps, logs, scale, ssh, stop, restart, run-task"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	LogCacheClient sharedaction.LogCacheClient
}

func (cmd *StartCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	logCacheEndpoint, _, err := cmd.Actor.GetLogCacheEndpoint()
	if err != nil {
		return err
	}
	cmd.LogCacheClient = command.NewLogCacheClient(logCacheEndpoint, config, ui)

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

	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	packageGuid, warnings, err := cmd.Actor.GetUnstagedNewestPackageGUID(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	if packageGuid != "" {
		cmd.UI.DisplayText("Staging app and tracing logs")

		logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.LogCacheClient)
		cmd.UI.DisplayWarnings(logWarnings)
		if logErr != nil {
			return logErr
		}
		defer stopLogStreamFunc()

		dropletStream, warningsStream, errStream := cmd.Actor.StagePackage(packageGuid, cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)

		droplet, err := shared.PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
		if err != nil {
			return err
		}

		warnings, err = cmd.Actor.SetApplicationDroplet(app.GUID, droplet.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayText("\nWaiting for app to start...")

	warnings, err = cmd.Actor.StartApplication(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	handleInstanceDetails := func(instanceDetails string) {
		cmd.UI.DisplayText(instanceDetails)
	}

	warnings, err = cmd.Actor.PollStart(app.GUID, false, handleInstanceDetails)
	cmd.UI.DisplayNewline()
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
