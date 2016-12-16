package v3

import (
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//These constants are only for filling in translations.
const (
	runningState   = "RUNNING"
	cancelingState = "CANCELING"
	pendingState   = "PENDING"
	succeededState = "SUCCEEDED"
)

//go:generate counterfeiter . TasksActor

type TasksActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationTasks(appGUID string, sortOrder v3action.SortOrder) ([]v3action.Task, v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

type TasksCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME tasks APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, logs, run-task, terminate-task"`

	UI     command.UI
	Actor  TasksActor
	Config command.Config
}

func (cmd *TasksCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client)

	return nil
}

func (cmd TasksCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Getting tasks for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	tasks, warnings, err := cmd.Actor.GetApplicationTasks(application.GUID, v3action.Descending)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	table := [][]string{{"id", "name", "state", "start time", "command"}}
	for _, task := range tasks {
		t, err := time.Parse(time.RFC3339, task.CreatedAt)
		if err != nil {
			return err
		}

		if task.Command == "" {
			task.Command = "[hidden]"
		}

		table = append(table, []string{
			strconv.Itoa(task.SequenceID),
			task.Name,
			cmd.UI.TranslateText(task.State),
			t.Format(time.RFC1123),
			task.Command,
		})
	}

	cmd.UI.DisplayTable("", table, 3)

	return nil
}
