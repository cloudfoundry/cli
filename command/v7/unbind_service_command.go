package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
)

type UnbindServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.BindServiceArgs `positional-args:"yes"`
	Wait            bool                 `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands interface{}          `related_commands:"apps, delete-service, services"`
}

func (cmd UnbindServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.DeleteServiceAppBinding(v7action.DeleteServiceAppBindingParams{
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		ServiceInstanceName: cmd.RequiredArgs.ServiceInstanceName,
		AppName:             cmd.RequiredArgs.AppName,
	})
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.ServiceBindingNotFoundError:
		cmd.UI.DisplayText("Binding between {{.ServiceInstanceName}} and {{.AppName}} did not exist", cmd.names())
		cmd.UI.DisplayOK()
		return nil
	default:
		return err
	}

	completed, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	switch {
	case err != nil:
		return err
	case completed:
		cmd.UI.DisplayOK()
		return nil
	default:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayText("Unbinding in progress. Use 'cf service {{.ServiceInstanceName}}' to check operation status.", cmd.names())
		return nil
	}
}

func (cmd UnbindServiceCommand) Usage() string {
	return `CF_NAME unbind-service APP_NAME SERVICE_INSTANCE`
}

func (cmd UnbindServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Unbinding app {{.AppName}} from service {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstanceName,
			"AppName":         cmd.RequiredArgs.AppName,
			"User":            user.Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"Org":             cmd.Config.TargetedOrganization().Name,
		},
	)

	return nil
}

func (cmd UnbindServiceCommand) names() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstanceName,
		"AppName":             cmd.RequiredArgs.AppName,
	}
}
