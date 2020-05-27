package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"
)

type CopySourceCommand struct {
	BaseCommand

	RequiredArgs        flag.CopySourceArgs     `positional-args:"yes"`
	usage               interface{}             `usage:"CF_NAME copy-source SOURCE_APP DESTINATION_APP [-s TARGET_SPACE [-o TARGET_ORG]] [--no-restart] [--strategy STRATEGY] [--no-wait]"`
	Strategy            flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy, either rolling or null"`
	NoWait              bool                    `long:"no-wait" description:"Exit when the first instance of the web process is healthy"`
	NoRestart           bool                    `long:"no-restart" description:"Do not restage the destination application"`
	Organization        string                  `short:"o" long:"organization" description:"Org that contains the destination application"`
	Space               string                  `short:"s" long:"space" description:"Space that contains the destination application"`
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

	return nil
}

func (cmd *CopySourceCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	logCacheEndpoint, _, err := cmd.Actor.GetLogCacheEndpoint()
	if err != nil {
		return err
	}
	logCacheClient := command.NewLogCacheClient(logCacheEndpoint, config, ui)
	cmd.Stager = shared.NewAppStager(cmd.Actor, cmd.UI, cmd.Config, logCacheClient)

	return nil
}

func (cmd CopySourceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
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

		err = cmd.Stager.StageAndStart(
			targetApp,
			targetSpace,
			targetOrg,
			pkg.GUID,
			cmd.Strategy.Name,
			cmd.NoWait,
			constant.ApplicationRestarting,
		)
		if err != nil {
			return mapErr(cmd.Config, targetApp.Name, err)
		}
	} else {
		cmd.UI.DisplayOK()
	}

	return nil
}
