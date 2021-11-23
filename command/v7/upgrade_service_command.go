package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UpgradeServiceCommand struct {
	BaseCommand

	RequiredArgs flag.ServiceInstance `positional-args:"yes"`
	Force        bool                 `short:"f" long:"force" description:"Force upgrade without asking for confirmation"`

	relatedCommands interface{} `related_commands:"services, update-service, update-user-provided-service"`
}

func (cmd UpgradeServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if !cmd.Force {
		upgrade, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !upgrade {
			cmd.UI.DisplayText("Upgrade cancelled")
			return nil
		}
	}

	if err := cmd.displayEvent(); err != nil {
		return err
	}

	serviceInstanceName := string(cmd.RequiredArgs.ServiceInstance)

	warnings, actorError := cmd.Actor.UpgradeManagedServiceInstance(
		serviceInstanceName,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)

	switch actorError.(type) {
	case nil:
		cmd.displayUpgradeInProgressMessage()
		cmd.UI.DisplayOK()
	case actionerror.ServiceInstanceUpgradeNotAvailableError:
		cmd.UI.DisplayText(actorError.Error())
		cmd.UI.DisplayOK()
	case actionerror.ServiceInstanceNotFoundError:
		return translatableerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return actorError
	}

	return nil
}

func (cmd UpgradeServiceCommand) Usage() string {
	return "CF_NAME upgrade-service SERVICE_INSTANCE"
}

func (cmd UpgradeServiceCommand) displayEvent() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Upgrading service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             cmd.Config.TargetedOrganization().Name,
			"SpaceName":           cmd.Config.TargetedSpace().Name,
			"Username":            user.Name,
		},
	)

	return nil
}

func (cmd UpgradeServiceCommand) displayPrompt() (bool, error) {
	cmd.UI.DisplayText(
		"Warning: This operation may be long running and will block further operations " +
			"on the service instance until it's completed",
	)

	upgrade, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Do you really want to upgrade the service instance {{.ServiceInstanceName}}?",
		cmd.serviceInstanceName(),
	)
	if err != nil {
		return false, err
	}

	return upgrade, nil
}

func (cmd UpgradeServiceCommand) displayUpgradeInProgressMessage() {
	cmd.UI.DisplayTextWithFlavor("Upgrade in progress. Use 'cf services' or 'cf service {{.ServiceInstance}}' to check operation status.",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		})
}

func (cmd UpgradeServiceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
