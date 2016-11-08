package v3

import (
	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v3/common"
)

//go:generate counterfeiter . RunTaskActor

type RunTaskActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3actions.Application, v3actions.Warnings, error)
	RunTask(appGUID string, command string) (v3actions.Task, v3actions.Warnings, error)
}

type RunTaskCommand struct {
	RequiredArgs    flags.RunTaskArgs `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME run-task APP_NAME COMMAND\n\nEXAMPLES:\n   CF_NAME run-task my-app \"bundle exec rake db:migrate\""`
	relatedCommands interface{}       `related_commands:"tasks, terminate-task"`

	UI     commands.UI
	Actor  RunTaskActor
	Config commands.Config
}

func (cmd *RunTaskCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := common.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3actions.NewActor(client)

	return nil
}

func (cmd RunTaskCommand) Execute(args []string) error {
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

	cmd.UI.DisplayHeaderFlavorText("Creating task for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	task, warnings, err := cmd.Actor.RunTask(application.GUID, cmd.RequiredArgs.Command)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return common.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Task {{.TaskSequenceID}} has been submitted successfully for execution.",
		map[string]interface{}{
			"TaskSequenceID": task.SequenceID,
		})

	return nil
}
