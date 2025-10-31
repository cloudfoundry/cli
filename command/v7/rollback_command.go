package v7

import (
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/sharedaction"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/cf/errors"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
    "code.cloudfoundry.org/cli/v9/resources"
)

type RollbackCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName            `positional-args:"yes"`
	Force           bool                    `short:"f" description:"Force rollback without confirmation"`
	InstanceSteps   string                  `long:"instance-steps" description:"An array of percentage steps to deploy when using deployment strategy canary. (e.g. 20,40,60)"`
	MaxInFlight     *int                    `long:"max-in-flight" description:"Defines the maximum number of instances that will be actively being rolled back."`
	Strategy        flag.DeploymentStrategy `long:"strategy" description:"Deployment strategy can be canary or rolling. When not specified, it defaults to rolling."`
	Version         flag.Revision           `long:"version" required:"true" description:"Roll back to the specified revision"`
	relatedCommands interface{}             `related_commands:"revision, revisions"`
	usage           interface{}             `usage:"CF_NAME rollback APP_NAME [--version VERSION] [-f]"`

	LogCacheClient sharedaction.LogCacheClient
	Stager         shared.AppStager
}

func (cmd *RollbackCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	cmd.Stager = shared.NewAppStager(cmd.Actor, cmd.UI, cmd.Config, cmd.LogCacheClient)
	return nil
}

func (cmd RollbackCommand) Execute(args []string) error {
	targetRevision := int(cmd.Version.Value)
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

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	revisions, warnings, err := cmd.Actor.GetRevisionsByApplicationNameAndSpace(app.Name, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(revisions) == 0 {
		return errors.New(fmt.Sprintf("No revisions for app %s", cmd.RequiredArgs.AppName))
	}

	revision, warnings, err := cmd.Actor.GetRevisionByApplicationAndVersion(app.GUID, targetRevision)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	// TODO Localization?
	if !cmd.Force {
		cmd.UI.DisplayTextWithFlavor("Rolling '{{.AppName}}' back to revision '{{.TargetRevision}}' will create a new revision. The new revision will use the settings from revision '{{.TargetRevision}}'.", map[string]interface{}{
			"AppName":        cmd.RequiredArgs.AppName,
			"TargetRevision": targetRevision,
		})

		prompt := "Are you sure you want to continue?"
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, nil)

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("App '{{.AppName}}' has not been rolled back to revision '{{.TargetRevision}}'.", map[string]interface{}{
				"AppName":        cmd.RequiredArgs.AppName,
				"TargetRevision": targetRevision,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Rolling back to revision {{.TargetRevision}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":        cmd.RequiredArgs.AppName,
		"TargetRevision": targetRevision,
		"OrgName":        cmd.Config.TargetedOrganization().Name,
		"SpaceName":      cmd.Config.TargetedSpace().Name,
		"Username":       user.Name,
	})

	opts := shared.AppStartOpts{
		AppAction: constant.ApplicationRollingBack,
		NoWait:    false,
		Strategy:  constant.DeploymentStrategyRolling,
	}
	if cmd.MaxInFlight != nil {
		opts.MaxInFlight = *cmd.MaxInFlight
	}

	if cmd.Strategy.Name != "" {
		opts.Strategy = cmd.Strategy.Name
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

	startAppErr := cmd.Stager.StartApp(app, cmd.Config.TargetedSpace(), cmd.Config.TargetedOrganization(), revision.GUID, opts)
	if startAppErr != nil {
		return startAppErr
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd RollbackCommand) ValidateFlags() error {
	switch {
	case cmd.MaxInFlight != nil && *cmd.MaxInFlight < 1:
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
