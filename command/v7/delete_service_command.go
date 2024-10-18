package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
)

type DeleteServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" long:"force" description:"Force deletion without confirmation"`
	Wait            bool                 `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands interface{}          `related_commands:"unbind-service, services"`
}

func (cmd DeleteServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if !cmd.Force {
		delete, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !delete {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	if err := cmd.displayEvent(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.DeleteServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)

	switch err.(type) {
	case nil:
	case actionerror.ServiceInstanceNotFoundError:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} did not exist.", cmd.serviceInstanceName())
		cmd.UI.DisplayOK()
		return nil
	default:
		return err
	}

	deleted, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	if err != nil {
		return err
	}

	switch deleted {
	case true:
		cmd.UI.DisplayTextWithFlavor("Service instance {{.ServiceInstanceName}} deleted.", cmd.serviceInstanceName())
	default:
		cmd.UI.DisplayText("Delete in progress. Use 'cf services' or 'cf service {{.ServiceInstanceName}}' to check operation status.", cmd.serviceInstanceName())
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd DeleteServiceCommand) Usage() string {
	return "CF_NAME delete-service SERVICE_INSTANCE [-f] [-w]"
}

func (cmd DeleteServiceCommand) displayEvent() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Deleting service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             cmd.Config.TargetedOrganization().Name,
			"SpaceName":           cmd.Config.TargetedSpace().Name,
			"Username":            user.Name,
		},
	)
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd DeleteServiceCommand) displayPrompt() (bool, error) {
	cmd.UI.DisplayText("This action impacts all resources scoped to this service instance, including service bindings, service keys and route bindings.")
	cmd.UI.DisplayText("This will remove the service instance from any spaces where it has been shared.")

	delete, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Really delete the service instance {{.ServiceInstanceName}}?",
		cmd.serviceInstanceName(),
	)
	if err != nil {
		return false, err
	}

	return delete, nil
}

func (cmd DeleteServiceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
