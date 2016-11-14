package v3

import (
	"net/url"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v3/common"
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
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3actions.Application, v3actions.Warnings, error)
	GetApplicationTasks(appGUID string, sortOrder v3actions.SortOrder) ([]v3actions.Task, v3actions.Warnings, error)
}

type TasksCommand struct {
	RequiredArgs    flags.AppName `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME tasks APP_NAME"`
	relatedCommands interface{}   `related_commands:"apps, run-task, terminate-task"`

	UI     commands.UI
	Actor  TasksActor
	Config commands.Config
}

func (cmd *TasksCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := common.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v3actions.NewActor(client)

	return nil
}

func (cmd TasksCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Getting tasks for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	query := url.Values{}
	query.Add("order_by", "-created_at")
	tasks, warnings, err := cmd.Actor.GetApplicationTasks(application.GUID, v3actions.Descending)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return common.HandleError(err)
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
