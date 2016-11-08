package v3

import (
	"fmt"

	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v3/common"
)

//go:generate counterfeiter . TasksActor

type TasksActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3actions.Application, v3actions.Warnings, error)
	GetApplicationTasks(appGUID string) ([]v3actions.Task, v3actions.Warnings, error)
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

	client, err := common.NewCloudControllerClient(config)
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

	cmd.UI.DisplayHeaderFlavorText("Getting tasks for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	tasks, warnings, err := cmd.Actor.GetApplicationTasks(application.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return common.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	// display tasks in table
	fmt.Println(tasks)

	return nil
}
