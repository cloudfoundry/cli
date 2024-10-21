package v6

import (
	"fmt"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/actor/sharedaction"
	"code.cloudfoundry.org/cli/v7/actor/v2action"
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/flag"
	"code.cloudfoundry.org/cli/v7/command/v6/shared"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UnbindServiceActor

type UnbindServiceActor interface {
	UnbindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string) (v2action.ServiceBinding, v2action.Warnings, error)
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

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
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

	serviceBinding, warnings, err := cmd.Actor.UnbindServiceBySpace(cmd.RequiredArgs.AppName, cmd.RequiredArgs.ServiceInstanceName, space.GUID)
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

	if serviceBinding.IsInProgress() {
		cmd.UI.DisplayText("Unbinding in progress. Use '{{.CFCommand}} {{.ServiceName}}' to check operation status.", map[string]interface{}{
			"CFCommand":   fmt.Sprintf("%s service", cmd.Config.BinaryName()),
			"ServiceName": cmd.RequiredArgs.ServiceInstanceName,
		})
	}

	return nil
}
