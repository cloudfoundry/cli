package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"fmt"
)

type CopySourceCommand struct {
	BaseCommand

	RequiredArgs        flag.CopySourceArgs `positional-args:"yes"`
	usage               interface{}         `usage:"CF_NAME copy-source SOURCE_APP DESTINATION_APP"`
	Organization        string              `short:"o" long:"organization" description:"Org that contains the destination application"`
	Space               string              `short:"s" long:"space" description:"Space that contains the destination application"`
	relatedCommands     interface{}         `related_commands:"apps, push, restage, restart, target"`
	envCFStagingTimeout interface{}         `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}         `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Stager shared.AppStager
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

	if cmd.Organization != "" && cmd.Space == "" {
		return errors.New("Incorrect Usage: '--organization, -o' requires '--space, -s' to be specified")
	}


	targetOrg := cmd.Config.TargetedOrganization()
	targetSpace := cmd.Config.TargetedSpace()
	fmt.Printf("0 --> target space: %s\n", targetSpace.Name)
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
		fmt.Printf("1 --> target space: %s\n", targetSpace.Name)

	}
	fmt.Printf("2 --> target space: %s\n", targetSpace.Name)


	cmd.UI.DisplayTextWithFlavor(
		"Copying source from app {{.SourceApp}} to target app {{.TargetApp}} in org {{.Org}} / space {{.Space}} as {{.UserName}}...",
		map[string]interface{}{
			"SourceApp": cmd.RequiredArgs.SourceAppName,
			"TargetApp": cmd.RequiredArgs.TargetAppName,
			"Org":       targetOrg.Name,
			"Space":     targetSpace.Name,
			"UserName":  user.Name,
		},
	)

	sourceApp, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.SourceAppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	fmt.Printf("3 --> target space: %s\n", targetSpace.Name)
	targetApp, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.TargetAppName, targetSpace.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	pkg, warnings, err := cmd.Actor.CopyPackage(sourceApp.GUID, targetApp.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	err = cmd.Stager.StageAndStart(
		targetApp,
		pkg.GUID,
		constant.DeploymentStrategyDefault,
		false,
	)
	if err != nil {
		return mapErr(cmd.Config, targetApp.Name, err)
	}

	return nil
}
