package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type TaskCommand struct {
	BaseCommand

	RequiredArgs    flag.TaskArgs `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME task APP_NAME TASK_ID"`
	relatedCommands interface{}   `related_commands:"apps, logs, run-task, tasks, terminate-task"`
}

func (cmd TaskCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	space := cmd.Config.TargetedSpace()

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	application, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting task {{.TaskID}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"TaskID":      cmd.RequiredArgs.TaskID,
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	task, warnings, err := cmd.Actor.GetTaskBySequenceIDAndApplication(cmd.RequiredArgs.TaskID, application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if task.Command == "" {
		task.Command = "[hidden]"
	}

	table := [][]string{
		{cmd.UI.TranslateText("id:"), strconv.FormatInt(task.SequenceID, 10)},
		{cmd.UI.TranslateText("name:"), task.Name},
		{cmd.UI.TranslateText("state:"), string(task.State)},
		{cmd.UI.TranslateText("start time:"), task.CreatedAt},
		{cmd.UI.TranslateText("command:"), task.Command},
		{cmd.UI.TranslateText("memory in mb:"), strconv.FormatUint(task.MemoryInMB, 10)},
		{cmd.UI.TranslateText("disk in mb:"), strconv.FormatUint(task.DiskInMB, 10)},
		{cmd.UI.TranslateText("log rate limit:"), strconv.Itoa(task.LogRateLimitInBPS)},
		{cmd.UI.TranslateText("failure reason:"), task.Result.FailureReason},
	}

	cmd.UI.DisplayKeyValueTable("", table, ui.DefaultTableSpacePadding)

	return nil
}
