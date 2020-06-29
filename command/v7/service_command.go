package v7

import "code.cloudfoundry.org/cli/command/flag"

type ServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	ShowGUID        bool                 `long:"guid" description:"Retrieve and display the given service's guid. All other output for the service is suppressed."`
	usage           interface{}          `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"bind-service, rename-service, update-service"`
}

func (cmd ServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	serviceInstance, warnings, err := cmd.Actor.GetServiceInstanceByNameAndSpace(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(serviceInstance.GUID)
	return nil
}
