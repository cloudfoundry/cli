package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/sharedaction"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/api/logcache"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
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

	cmd.LogCacheClient, err = logcache.NewClient(config.LogCacheEndpoint(), config, ui, v7action.NewDefaultKubernetesConfigGetter())
	if err != nil {
		return err
	}

	cmd.Stager = shared.NewAppStager(cmd.Actor, cmd.UI, cmd.Config, cmd.LogCacheClient)

	return nil
}

func (cmd StartCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
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

	opts := shared.AppStartOpts{
		Strategy:  constant.DeploymentStrategyDefault,
		NoWait:    false,
		AppAction: constant.ApplicationStarting,
	}

	if packageGUID != "" && app.Stopped() {
		cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
		cmd.UI.DisplayNewline()

		return cmd.Stager.StageAndStart(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), packageGUID, opts)
	}

	return cmd.Stager.StartApp(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), "", opts)
}
