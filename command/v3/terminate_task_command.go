package v3

import (
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . TerminateTaskActor

type TerminateTaskActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetTaskBySequenceIDAndApplication(sequenceID int, appGUID string) (v3action.Task, v3action.Warnings, error)
	TerminateTask(taskGUID string) (v3action.Task, v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

type TerminateTaskCommand struct {
	RequiredArgs    flag.TerminateTaskArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME terminate-task APP_NAME TASK_ID\n\nEXAMPLES:\n   CF_NAME terminate-task my-app 3"`
	relatedCommands interface{}            `related_commands:"tasks"`

	UI     command.UI
	Actor  TerminateTaskActor
	Config command.Config
}

func (cmd *TerminateTaskCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client)

	return nil
}

func (cmd TerminateTaskCommand) Execute(args []string) error {
	sequenceId, err := flag.ParseStringToInt(cmd.RequiredArgs.SequenceID)
	if err != nil {
		return command.ParseArgumentError{
			ArgumentName: "TASK_ID",
			ExpectedType: "integer",
		}
	}

	err = command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), "3.0.0")
	if err != nil {
		return err
	}

	err = command.CheckTarget(cmd.Config, true, true)
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
		return shared.HandleError(err)
	}

	task, warnings, err := cmd.Actor.GetTaskBySequenceIDAndApplication(sequenceId, application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
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
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
