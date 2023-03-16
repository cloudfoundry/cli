package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type RollbackCommand struct {
	BaseCommand

	Force           bool          `short:"f" description:"Force rollback without confirmation"`
	RequiredArgs    flag.AppName  `positional-args:"yes"`
	Version         flag.Revision `long:"version" required:"true" description:"Roll back to the specified revision"`
	relatedCommands interface{}   `related_commands:"revisions"`
	usage           interface{}   `usage:"CF_NAME rollback APP_NAME [--version VERSION] [-f]"`

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
	cmd.UI.DisplayWarning(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	targetRevision := int(cmd.Version.Value)
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

	startAppErr := cmd.Stager.StartApp(
		app,
		revision.GUID,
		constant.DeploymentStrategyRolling,
		false,
		cmd.Config.TargetedSpace(),
		cmd.Config.TargetedOrganization(),
		constant.ApplicationRollingBack,
	)
	if startAppErr != nil {
		return startAppErr
	}

	cmd.UI.DisplayOK()

	return nil
}
