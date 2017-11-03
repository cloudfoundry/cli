package v2

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . UnbindServiceActor

type UnbindServiceActor interface {
	UnbindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string) (v2action.Warnings, error)
}

type UnbindServiceCommand struct {
	RequiredArgs    flag.BindServiceArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME unbind-service APP_NAME SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"apps, delete-service, services"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UnbindServiceActor
}

func (cmd *UnbindServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd UnbindServiceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	space := cmd.Config.TargetedSpace()
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Unbinding app {{.AppName}} from service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":     cmd.RequiredArgs.AppName,
		"ServiceName": cmd.RequiredArgs.ServiceInstanceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   space.Name,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.UnbindServiceBySpace(cmd.RequiredArgs.AppName, cmd.RequiredArgs.ServiceInstanceName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.ServiceBindingNotFoundError); ok {
			cmd.UI.DisplayWarning("Binding between {{.InstanceName}} and {{.AppName}} did not exist", map[string]interface{}{
				"AppName":      cmd.RequiredArgs.AppName,
				"InstanceName": cmd.RequiredArgs.ServiceInstanceName,
			})
		} else {
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
