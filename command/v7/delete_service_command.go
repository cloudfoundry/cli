package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" long:"force" description:"Force deletion without confirmation"`
	Wait            bool                 `short:"w" long:"wait" description:"Wait for the delete operation to complete"`
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

	state, warnings, err := cmd.Actor.DeleteServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
		cmd.Wait,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayState(state)
	return nil
}

func (cmd DeleteServiceCommand) Usage() string {
	return "CF_NAME delete-service SERVICE_INSTANCE [-f] [-w]"
}

func (cmd DeleteServiceCommand) displayEvent() error {
	user, err := cmd.Config.CurrentUser()
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

	return nil
}

func (cmd DeleteServiceCommand) displayPrompt() (bool, error) {
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

func (cmd DeleteServiceCommand) displayState(state v7action.ServiceInstanceDeleteState) {
	cmd.UI.DisplayNewline()
	switch state {
	case v7action.ServiceInstanceDidNotExist:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} did not exist.", cmd.serviceInstanceName())
	case v7action.ServiceInstanceGone:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} deleted.", cmd.serviceInstanceName())
	case v7action.ServiceInstanceDeleteInProgress:
		cmd.UI.DisplayText("Delete in progress. Use 'cf services' or 'cf service {{.ServiceInstanceName}}' to check operation status.", cmd.serviceInstanceName())
	}
	cmd.UI.DisplayOK()
}

func (cmd DeleteServiceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
