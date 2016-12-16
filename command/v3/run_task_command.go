package v3

import (
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . RunTaskActor

type RunTaskActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	RunTask(appGUID string, command string, name string) (v3action.Task, v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

type RunTaskCommand struct {
	RequiredArgs    flag.RunTaskArgs `positional-args:"yes"`
	Name            string           `long:"name" description:"Name to give the task (generated if omitted)"`
	usage           interface{}      `usage:"CF_NAME run-task APP_NAME COMMAND [--name TASK_NAME]\n\nEXAMPLES:\n   CF_NAME run-task my-app \"bundle exec rake db:migrate\" --name migrate"`
	relatedCommands interface{}      `related_commands:"logs, tasks, terminate-task"`

	UI     command.UI
	Actor  RunTaskActor
	Config command.Config
}

func (cmd *RunTaskCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client)

	return nil
}

func (cmd RunTaskCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), "3.0.0")
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

	cmd.UI.DisplayTextWithFlavor("Creating task for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	task, warnings, err := cmd.Actor.RunTask(application.GUID, cmd.RequiredArgs.Command, cmd.Name)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Task {{.TaskSequenceID}} has been submitted successfully for execution.",
		map[string]interface{}{
			"TaskSequenceID": task.SequenceID,
		})

	return nil
}
