package v7

import (
    "strconv"
    "strings"

    "code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
    "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/api/logcache"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
    "code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
)

type CopySourceCommand struct {
	BaseCommand

	RequiredArgs        flag.CopySourceArgs     `positional-args:"yes"`
	usage               interface{}             `usage:"CF_NAME copy-source SOURCE_APP DESTINATION_APP [-s TARGET_SPACE [-o TARGET_ORG]] [--no-restart] [--strategy STRATEGY] [--no-wait]"`
	InstanceSteps       string                  `long:"instance-steps" description:"An array of percentage steps to deploy when using deployment strategy canary. (e.g. 20,40,60)"`
	MaxInFlight         *int                    `long:"max-in-flight" description:"Defines the maximum number of instances that will be actively being started. Only applies when --strategy flag is specified."`
	NoWait              bool                    `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	NoRestart           bool                    `long:"no-restart" description:"Do not restage the destination application"`
	Organization        string                  `short:"o" long:"organization" description:"Org that contains the destination application"`
	Space               string                  `short:"s" long:"space" description:"Space that contains the destination application"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy can be canary, rolling or null"`
	relatedCommands     interface{}             `related_commands:"apps, push, restage, restart, target"`
	envCFStagingTimeout interface{}             `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Stager shared.AppStager
}

func (cmd *CopySourceCommand) ValidateFlags() error {
	if cmd.Organization != "" && cmd.Space == "" {
		return translatableerror.RequiredFlagsError{
			Arg1: "--organization, -o",
			Arg2: "--space, -s",
		}
	}

	if cmd.NoRestart && cmd.Strategy.Name != constant.DeploymentStrategyDefault {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--no-restart", "--strategy"},
		}
	}

	if cmd.NoRestart && cmd.NoWait {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--no-restart", "--no-wait"},
		}
	}

	if cmd.Strategy.Name == constant.DeploymentStrategyDefault && cmd.MaxInFlight != nil {
		return translatableerror.RequiredFlagsError{Arg1: "--max-in-flight", Arg2: "--strategy"}
	}

	if cmd.Strategy.Name != constant.DeploymentStrategyDefault && cmd.MaxInFlight != nil && *cmd.MaxInFlight < 1 {
		return translatableerror.IncorrectUsageError{Message: "--max-in-flight must be greater than or equal to 1"}
	}

	if cmd.Strategy.Name != constant.DeploymentStrategyCanary && cmd.InstanceSteps != "" {
		return translatableerror.RequiredFlagsError{Arg1: "--instance-steps", Arg2: "--strategy=canary"}
	}

	if len(cmd.InstanceSteps) > 0 && !validateInstanceSteps(cmd.InstanceSteps) {
		return translatableerror.ParseArgumentError{ArgumentName: "--instance-steps", ExpectedType: "list of weights"}
	}

	if len(cmd.InstanceSteps) > 0 {
		return command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionCanarySteps, "--instance-steps")
	}

	return nil
}

func (cmd *CopySourceCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CopySourceCommand) Execute(args []string) error {
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

	targetOrgName := cmd.Config.TargetedOrganization().Name
	targetSpaceName := cmd.Config.TargetedSpace().Name
	if cmd.Organization != "" {
		targetOrgName = cmd.Organization
	}
	if cmd.Space != "" {
		targetSpaceName = cmd.Space
	}

	cmd.UI.DisplayTextWithFlavor(
		"Copying source from app {{.SourceApp}} to target app {{.TargetApp}} in org {{.Org}} / space {{.Space}} as {{.UserName}}...",
		map[string]interface{}{
			"SourceApp": cmd.RequiredArgs.SourceAppName,
			"TargetApp": cmd.RequiredArgs.TargetAppName,
			"Org":       targetOrgName,
			"Space":     targetSpaceName,
			"UserName":  user.Name,
		},
	)
	cmd.UI.DisplayNewline()

	targetOrg := cmd.Config.TargetedOrganization()
	targetSpace := cmd.Config.TargetedSpace()
	if cmd.Organization != "" {
		org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		targetOrg = configv3.Organization{
			GUID: org.GUID,
			Name: org.Name,
		}
	}
	if cmd.Space != "" {
		space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, targetOrg.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		targetSpace = configv3.Space{
			GUID: space.GUID,
			Name: space.Name,
		}
	}

	sourceApp, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.SourceAppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	targetApp, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.TargetAppName, targetSpace.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	pkg, warnings, err := cmd.Actor.CopyPackage(sourceApp, targetApp)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if !cmd.NoRestart {
		cmd.UI.DisplayTextWithFlavor(
			"Staging app {{.TargetApp}} in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.UserName}}...",
			map[string]interface{}{
				"TargetApp":   cmd.RequiredArgs.TargetAppName,
				"TargetOrg":   targetOrgName,
				"TargetSpace": targetSpaceName,
				"UserName":    user.Name,
			},
		)
		cmd.UI.DisplayNewline()

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

		err = cmd.Stager.StageAndStart(targetApp, targetSpace, targetOrg, pkg.GUID, opts)
		if err != nil {
			return mapErr(cmd.Config, targetApp.Name, err)
		}
	} else {
		cmd.UI.DisplayOK()
	}

	return nil
}
