package v7

import (
    "strconv"
    "strings"

    "code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/api/logcache"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
    "code.cloudfoundry.org/cli/v9/resources"
)

type RestageCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName            `positional-args:"yes"`
	InstanceSteps       string                  `long:"instance-steps" description:"An array of percentage steps to deploy when using deployment strategy canary. (e.g. 20,40,60)"`
	MaxInFlight         *int                    `long:"max-in-flight" description:"Defines the maximum number of instances that will be actively being restaged. Only applies when --strategy flag is specified."`
	NoWait              bool                    `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy can be canary, rolling or null."`
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

	err = cmd.ValidateFlags()
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
		AppAction: constant.ApplicationRestarting,
		NoWait:    cmd.NoWait,
		Strategy:  cmd.Strategy.Name,
	}

	if cmd.MaxInFlight != nil {
		opts.MaxInFlight = *cmd.MaxInFlight
	}

	if cmd.InstanceSteps != "" {
		if len(cmd.InstanceSteps) > 0 {
			for _, v := range strings.Split(cmd.InstanceSteps, ",") {
				parsedInt, err := strconv.ParseInt(v, 0, 64)
				if err != nil {
					return err
				}
				opts.CanarySteps = append(opts.CanarySteps, resources.CanaryStep{InstanceWeight: parsedInt})
			}
		}
	}

	err = cmd.Stager.StageAndStart(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), pkg.GUID, opts)
	if err != nil {
		return mapErr(cmd.Config, cmd.RequiredArgs.AppName, err)
	}

	return nil
}

func (cmd RestageCommand) ValidateFlags() error {
	switch {
	case cmd.Strategy.Name == constant.DeploymentStrategyDefault && cmd.MaxInFlight != nil:
		return translatableerror.RequiredFlagsError{Arg1: "--max-in-flight", Arg2: "--strategy"}
	case cmd.Strategy.Name != constant.DeploymentStrategyDefault && cmd.MaxInFlight != nil && *cmd.MaxInFlight < 1:
		return translatableerror.IncorrectUsageError{Message: "--max-in-flight must be greater than or equal to 1"}
	case cmd.Strategy.Name != constant.DeploymentStrategyCanary && cmd.InstanceSteps != "":
		return translatableerror.RequiredFlagsError{Arg1: "--instance-steps", Arg2: "--strategy=canary"}
	case len(cmd.InstanceSteps) > 0 && !validateInstanceSteps(cmd.InstanceSteps):
		return translatableerror.ParseArgumentError{ArgumentName: "--instance-steps", ExpectedType: "list of weights"}
	}

	if len(cmd.InstanceSteps) > 0 {
		return command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionCanarySteps, "--instance-steps")
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
