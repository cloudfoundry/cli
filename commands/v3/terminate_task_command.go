package v3

import (
	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v3/common"
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

	UI     commands.UI
	Actor  TerminateTaskActor
	Config commands.Config
}

func (cmd *TerminateTaskCommand) Setup(config commands.Config, ui commands.UI) error {
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
	err := common.CheckTarget(cmd.Config, true, true)
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

	task, warnings, err := cmd.Actor.GetTaskBySequenceIDAndApplication(cmd.RequiredArgs.SequenceID, application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return common.HandleError(err)
	}

	cmd.UI.DisplayHeaderFlavorText("Terminating task {{.TaskSequenceID}} of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
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
