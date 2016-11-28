package v3

import (
	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
	"code.cloudfoundry.org/cli/command/v3/common"
)

//go:generate counterfeiter . TerminateTaskActor

type TerminateTaskActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3actions.Application, v3actions.Warnings, error)
	GetTaskBySequenceIDAndApplication(sequenceID int, appGUID string) (v3actions.Task, v3actions.Warnings, error)
	TerminateTask(taskGUID string) (v3actions.Task, v3actions.Warnings, error)
}

type TerminateTaskCommand struct {
	RequiredArgs    flags.TerminateTaskArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME terminate-task APP_NAME TASK_ID\n\nEXAMPLES:\n   CF_NAME terminate-task my-app 3"`
	relatedCommands interface{}             `related_commands:"tasks"`

	UI     command.UI
	Actor  TerminateTaskActor
	Config command.Config
}

func (cmd *TerminateTaskCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := common.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3actions.NewActor(client)

	return nil
}

func (cmd TerminateTaskCommand) Execute(args []string) error {
	sequenceId, err := flags.ParseStringToInt(cmd.RequiredArgs.SequenceID)
	if err != nil {
		return common.ParseArgumentError{
			ArgumentName: "TASK_ID",
			ExpectedType: "integer",
		}
	}

	err = common.CheckTarget(cmd.Config, true, true)
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
		return common.HandleError(err)
	}

	task, warnings, err := cmd.Actor.GetTaskBySequenceIDAndApplication(sequenceId, application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return common.HandleError(err)
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
		return common.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
