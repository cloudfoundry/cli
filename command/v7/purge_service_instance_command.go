package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
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

	warnings, err := cmd.Actor.PurgeServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	cmd.UI.DisplayNewline()

	switch err.(type) {
	case nil:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} purged.", cmd.serviceInstanceName())
	case actionerror.ServiceInstanceNotFoundError:
		cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} did not exist.", cmd.serviceInstanceName())
	default:
		return err
	}

	cmd.UI.DisplayOK()
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
	user, err := cmd.Actor.GetCurrentUser()
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

func (cmd PurgeServiceInstanceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
