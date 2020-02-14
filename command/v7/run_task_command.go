package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/flag"
)

type RunTaskCommand struct {
	BaseCommand

	RequiredArgs    flag.RunTaskArgsV7 `positional-args:"yes"`
	Command         string             `long:"command" short:"c" description:"The command to execute"`
	Disk            flag.Megabytes     `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	Memory          flag.Megabytes     `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	Name            string             `long:"name" description:"Name to give the task (generated if omitted)"`
	Process         string             `long:"process" description:"Process type to use as a template for command, memory, and disk for the created task."`
	usage           interface{}        `usage:"CF_NAME run-task APP_NAME [--command COMMAND] [-k DISK] [-m MEMORY] [--name TASK_NAME] [--process PROCESS_TYPE]\n\nTIP:\n   Use 'cf logs' to display the logs of the app and all its tasks. If your task name is unique, grep this command's output for the task name to view task-specific logs.\n\nEXAMPLES:\n   CF_NAME run-task my-app --command \"bundle exec rake db:migrate\" --name migrate\n\n   CF_NAME run-task my-app --process batch_job\n\n   CF_NAME run-task my-app"`
	relatedCommands interface{}        `related_commands:"logs, tasks, terminate-task"`
}

func (cmd RunTaskCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
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

	inputTask := v7action.Task{
		Command: cmd.Command,
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
	if cmd.Command == "" && cmd.Process == "" {
		cmd.Process = "task"
	}
	if cmd.Process != "" {
		process, warnings, err := cmd.Actor.GetProcessByTypeAndApplication(cmd.Process, application.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		inputTask.Template = &ccv3.TaskTemplate{
			Process: ccv3.TaskProcessTemplate{
				Guid: process.GUID,
			},
		}
	}

	task, warnings, err := cmd.Actor.RunTask(application.GUID, inputTask)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Task has been submitted successfully for execution.")
	cmd.UI.DisplayOK()

	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("task name:"), task.Name},
		{cmd.UI.TranslateText("task id:"), fmt.Sprint(task.SequenceID)},
	}, 3)

	return nil
}
