package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/logcache"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type RestartCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName            `positional-args:"yes"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy can be canary, rolling or null."`
	NoWait              bool                    `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	usage               interface{}             `usage:"CF_NAME restart APP_NAME\n\n   This command will cause downtime unless you use '--strategy canary' or '--strategy rolling'.\n\n   If the app's most recent package is unstaged, restarting the app will stage and run that package.\n   Otherwise, the app's current droplet will be run."`
	relatedCommands     interface{}             `related_commands:"restage, restart-app-instance"`
	envCFStagingTimeout interface{}             `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Stager shared.AppStager
}

func (cmd *RestartCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	logCacheClient, err := logcache.NewClient(config.LogCacheEndpoint(), config, ui, v7action.NewDefaultKubernetesConfigGetter())
	if err != nil {
		return err
	}

	cmd.Stager = shared.NewAppStager(cmd.Actor, cmd.UI, cmd.Config, logCacheClient)

	return nil
}

func (cmd RestartCommand) Execute(args []string) error {
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

	if packageGUID != "" || len(cmd.Strategy.Name) > 0 {
		cmd.UI.DisplayTextWithFlavor("Restarting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
		cmd.UI.DisplayNewline()
	}

	opts := shared.AppStartOpts{
		Strategy:  cmd.Strategy.Name,
		NoWait:    cmd.NoWait,
		AppAction: constant.ApplicationRestarting,
	}
	if packageGUID != "" {
		err = cmd.Stager.StageAndStart(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), packageGUID, opts)
		if err != nil {
			return err
		}
	} else {
		err = cmd.Stager.StartApp(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), "", opts)
		if err != nil {
			return err
		}
	}

	return nil
}
