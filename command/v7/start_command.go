package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type StartCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME\n\n   If the app's most recent package is unstaged, starting the app will stage and run that package.\n   Otherwise, the app's current droplet will be run."`
	relatedCommands     interface{}  `related_commands:"apps, logs, scale, ssh, stop, restart, run-task"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	LogCacheClient sharedaction.LogCacheClient
	Stager         shared.AppStager
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
	cmd.Stager = shared.NewAppStager(cmd.Actor, cmd.UI, cmd.Config, cmd.LogCacheClient)

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

	packageGUID, warnings, err := cmd.Actor.GetUnstagedNewestPackageGUID(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if packageGUID != "" && app.Stopped() {
		cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
		cmd.UI.DisplayNewline()

		err = cmd.Stager.StageAndStart(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), packageGUID, constant.DeploymentStrategyDefault, false, constant.ApplicationStarting)
		if err != nil {
			return err
		}
	} else {
		err = cmd.Stager.StartApp(app, "", constant.DeploymentStrategyDefault, false, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), constant.ApplicationStarting)
		if err != nil {
			return err
		}
	}

	return nil

}
