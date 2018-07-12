package v3

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . RunTaskActor

type RunTaskActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	RunTask(appGUID string, task v3action.Task) (v3action.Task, v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

type RunTaskCommand struct {
	RequiredArgs    flag.RunTaskArgs `positional-args:"yes"`
	Disk            flag.Megabytes   `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	Memory          flag.Megabytes   `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	Name            string           `long:"name" description:"Name to give the task (generated if omitted)"`
	usage           interface{}      `usage:"CF_NAME run-task APP_NAME COMMAND [-k DISK] [-m MEMORY] [--name TASK_NAME]\n\nTIP:\n   Use 'cf logs' to display the logs of the app and all its tasks. If your task name is unique, grep this command's output for the task name to view task-specific logs.\n\nEXAMPLES:\n   CF_NAME run-task my-app \"bundle exec rake db:migrate\" --name migrate"`
	relatedCommands interface{}      `related_commands:"logs, tasks, terminate-task"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       RunTaskActor
}

func (cmd *RunTaskCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionRunTaskV3}
		}

		return err
	}
	cmd.Actor = v3action.NewActor(client, config, nil, nil)

	return nil
}

func (cmd RunTaskCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionRunTaskV3)
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

	cmd.UI.DisplayTextWithFlavor("Creating task for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	inputTask := v3action.Task{
		Command: cmd.RequiredArgs.Command,
	}

	if cmd.Name != "" {
		inputTask.Name = cmd.Name
	}
	if cmd.Disk.IsSet {
		inputTask.DiskInMB = cmd.Disk.Value
	}
	if cmd.Memory.IsSet {
		inputTask.MemoryInMB = cmd.Memory.Value
	}

	task, warnings, err := cmd.Actor.RunTask(application.GUID, inputTask)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Task has been submitted successfully for execution.")
	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("task name:"), task.Name},
		{cmd.UI.TranslateText("task id:"), fmt.Sprint(task.SequenceID)},
	}, 3)

	return nil
}
