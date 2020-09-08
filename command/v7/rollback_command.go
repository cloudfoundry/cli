package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RollbackCommand struct {
	BaseCommand

	Force           bool                 `short:"f" description:"Force rollback without confirmation"`
	RequiredArgs    flag.AppName         `positional-args:"yes"`
	Version         flag.PositiveInteger `long:"revision" required:"true" description:"Roll back to the given app revision"`
	relatedCommands interface{}          `related_commands:"revisions"`
	usage           interface{}          `usage:"CF_NAME rollback APP_NAME [--revision REVISION_NUMBER] [-f]"`
}

func (cmd RollbackCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	targetRevision := int(cmd.Version.Value)
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	app, warnings, _ := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)

	revisions, warnings, _ := cmd.Actor.GetRevisionsByApplicationNameAndSpace(app.Name, cmd.Config.TargetedSpace().GUID)

	cmd.UI.DisplayWarnings(warnings)
	if len(revisions) == 0 {
		return errors.New(fmt.Sprintf("No revisions for app %s", cmd.RequiredArgs.AppName))
	}
	newRevision := revisions[0].Version + 1
	revision, warnings, _ := cmd.Actor.GetRevisionByApplicationAndVersion(app.GUID, targetRevision)
	cmd.UI.DisplayWarnings(warnings)

	// TODO Localization?

	if !cmd.Force {
		cmd.UI.DisplayTextWithFlavor("Rolling '{{.AppName}}' back to revision '{{.TargetRevision}}' will create a new revision. The new revision '{{.NewRevision}}' will use the settings from revision '{{.TargetRevision}}'.", map[string]interface{}{
			"AppName":        cmd.RequiredArgs.AppName,
			"TargetRevision": targetRevision,
			"NewRevision":    newRevision,
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
	cmd.UI.DisplayNewline()

	_, warnings, _ = cmd.Actor.CreateDeploymentByApplicationAndRevision(app.GUID, revision.GUID)
	cmd.UI.DisplayWarnings(warnings)

	cmd.UI.DisplayOK()
	return nil
}
