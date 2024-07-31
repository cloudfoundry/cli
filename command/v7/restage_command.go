package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/logcache"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type RestageCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName            `positional-args:"yes"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy can be canary, rolling or null."`
	NoWait              bool                    `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	usage               interface{}             `usage:"CF_NAME restage APP_NAME\n\n   This command will cause downtime unless you use '--strategy' flag.\n\nEXAMPLES:\n   CF_NAME restage APP_NAME\n   CF_NAME restage APP_NAME --strategy rolling\n   CF_NAME restage APP_NAME --strategy canary --no-wait"`
	relatedCommands     interface{}             `related_commands:"restart"`
	envCFStagingTimeout interface{}             `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Stager shared.AppStager
}

func (cmd *RestageCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd RestageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	if len(cmd.Strategy.Name) <= 0 {
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

	pkg, warnings, err := cmd.Actor.GetNewestReadyPackageForApplication(app)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return mapErr(cmd.Config, cmd.RequiredArgs.AppName, err)
	}

	opts := shared.AppStartOpts{
		Strategy:  cmd.Strategy.Name,
		NoWait:    cmd.NoWait,
		AppAction: constant.ApplicationRestarting,
	}
	err = cmd.Stager.StageAndStart(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), pkg.GUID, opts)
	if err != nil {
		return mapErr(cmd.Config, cmd.RequiredArgs.AppName, err)
	}

	return nil
}

func mapErr(config command.Config, appName string, err error) error {
	switch err.(type) {
	case actionerror.AllInstancesCrashedError:
		return translatableerror.ApplicationUnableToStartError{
			AppName:    appName,
			BinaryName: config.BinaryName(),
		}
	case actionerror.StartupTimeoutError:
		return translatableerror.StartupTimeoutError{
			AppName:    appName,
			BinaryName: config.BinaryName(),
		}
	case actionerror.StagingFailedNoAppDetectedError:
		return translatableerror.StagingFailedNoAppDetectedError{
			Message:    err.Error(),
			BinaryName: config.BinaryName(),
		}
	case actionerror.NoEligiblePackagesError:
		return translatableerror.NoEligiblePackagesError{
			AppName:    appName,
			BinaryName: config.BinaryName(),
		}
	}
	return err
}
