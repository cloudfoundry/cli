package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type TerminateTaskCommand struct {
	command.BaseCommand

	RequiredArgs    flag.TerminateTaskArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME terminate-task APP_NAME TASK_ID\n\nEXAMPLES:\n   CF_NAME terminate-task my-app 3"`
	relatedCommands interface{}            `related_commands:"tasks"`
}

func (cmd TerminateTaskCommand) Execute(args []string) error {
	sequenceID, err := flag.ParseStringToInt(cmd.RequiredArgs.SequenceID)
	if err != nil {
		return translatableerror.ParseArgumentError{
			ArgumentName: "TASK_ID",
			ExpectedType: "integer",
		}
	}

	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	space := cmd.Config.TargetedSpace()

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	application, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	task, warnings, err := cmd.Actor.GetTaskBySequenceIDAndApplication(sequenceID, application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Terminating task {{.TaskSequenceID}} of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TaskSequenceID": cmd.RequiredArgs.SequenceID,
			"AppName":        cmd.RequiredArgs.AppName,
			"OrgName":        cmd.Config.TargetedOrganization().Name,
			"SpaceName":      space.Name,
			"CurrentUser":    user.Name,
		})

	_, warnings, err = cmd.Actor.TerminateTask(task.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
