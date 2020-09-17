package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"

	"code.cloudfoundry.org/cli/command/flag"
)

const purgeServiceInstanceWarning = "WARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys."

type PurgeServiceInstanceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" long:"force" description:"Force deletion without confirmation"`
	relatedCommands interface{}          `related_commands:"delete-service, service-brokers, services"`
}

func (cmd PurgeServiceInstanceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if !cmd.Force {
		purge, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !purge {
			cmd.UI.DisplayText("Purge cancelled")
			return nil
		}
	}

	if err := cmd.displayEvent(); err != nil {
		return err
	}

	state, warnings, err := cmd.Actor.PurgeServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayState(state)
	return nil
}

func (cmd PurgeServiceInstanceCommand) Usage() string {
	return strings.Join([]string{
		"CF_NAME purge-service-instance SERVICE_INSTANCE [-f]",
		"",
		purgeServiceInstanceWarning,
	}, "\n")
}

func (cmd PurgeServiceInstanceCommand) displayEvent() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Purging service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             cmd.Config.TargetedOrganization().Name,
			"SpaceName":           cmd.Config.TargetedSpace().Name,
			"Username":            user.Name,
		},
	)

	return nil
}

func (cmd PurgeServiceInstanceCommand) displayPrompt() (bool, error) {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText(purgeServiceInstanceWarning)
	cmd.UI.DisplayNewline()
	delete, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Really purge service instance {{.ServiceInstanceName}} from Cloud Foundry?",
		cmd.serviceInstanceName(),
	)
	if err != nil {
		return false, err
	}

	return delete, nil
}

func (cmd PurgeServiceInstanceCommand) displayState(state v7action.ServiceInstanceDeleteState) {
	cmd.UI.DisplayNewline()
	switch state {
	case v7action.ServiceInstanceDidNotExist:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} did not exist.", cmd.serviceInstanceName())
	default:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} purged.", cmd.serviceInstanceName())
	}
	cmd.UI.DisplayOK()
}

func (cmd PurgeServiceInstanceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
