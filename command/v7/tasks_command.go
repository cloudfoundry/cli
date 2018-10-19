package v7

import (
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
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
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetApplicationTasks(appGUID string, sortOrder v7action.SortOrder) ([]v7action.Task, v7action.Warnings, error)
}

type TasksCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME tasks APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, logs, run-task, terminate-task"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       TasksActor
}

func (cmd *TasksCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(client, config, nil, nil)

	return nil
}

func (cmd TasksCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Getting tasks for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	tasks, warnings, err := cmd.Actor.GetApplicationTasks(application.GUID, v7action.Descending)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	table := [][]string{
		{
			cmd.UI.TranslateText("id"),
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("state"),
			cmd.UI.TranslateText("start time"),
			cmd.UI.TranslateText("command"),
		},
	}
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
			cmd.UI.TranslateText(string(task.State)),
			t.Format(time.RFC1123),
			task.Command,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
