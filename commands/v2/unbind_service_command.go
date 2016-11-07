package v2

import (
	"os"

	"code.cloudfoundry.org/cli/actors/v2actions"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v2/common"
)

//go:generate counterfeiter . UnbindServiceActor

type UnbindServiceActor interface {
	UnbindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string) (v2actions.Warnings, error)
}

type UnbindServiceCommand struct {
	RequiredArgs    flags.BindServiceArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME unbind-service APP_NAME SERVICE_INSTANCE"`
	relatedCommands interface{}           `related_commands:"apps, delete-service, services"`

	UI     commands.UI
	Actor  UnbindServiceActor
	Config commands.Config
}

func (cmd *UnbindServiceCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, _, err := common.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2actions.NewActor(client)

	return nil
}

func (cmd UnbindServiceCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText(ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := common.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return err
	}

	space := cmd.Config.TargetedSpace()
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayHeaderFlavorText("Unbinding app {{.AppName}} from service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"ServiceName": cmd.RequiredArgs.ServiceInstanceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.UnbindServiceBySpace(cmd.RequiredArgs.AppName, cmd.RequiredArgs.ServiceInstanceName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(v2actions.ServiceBindingNotFoundError); ok {
			cmd.UI.DisplayWarning("Binding between {{.InstanceName}} and {{.AppName}} did not exist", map[string]interface{}{
				"AppName":      cmd.RequiredArgs.AppName,
				"InstanceName": cmd.RequiredArgs.ServiceInstanceName,
			})
		} else {
			return common.HandleError(err)
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
