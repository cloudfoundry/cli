package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
)

type ServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	ShowGUID        bool                 `long:"guid" description:"Retrieve and display the given service's guid. All other output for the service is suppressed."`
	usage           interface{}          `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"bind-service, rename-service, update-service"`

	serviceInstance resources.ServiceInstance
}

func (cmd ServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	var (
		warnings v7action.Warnings
		err      error
	)

	cmd.serviceInstance, warnings, err = cmd.Actor.GetServiceInstanceByNameAndSpace(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	switch {
	case cmd.ShowGUID:
		return cmd.displayGUID()
	case cmd.serviceInstance.Type == resources.UserProvidedServiceInstance:
		return cmd.displayUPSI()
	default:
		return nil
	}
}

func (cmd ServiceCommand) displayGUID() error {
	cmd.UI.DisplayText(cmd.serviceInstance.GUID)
	return nil
}

func (cmd ServiceCommand) displayUPSI() error {
	if err := cmd.displayInto(); err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), cmd.serviceInstance.Name},
		{cmd.UI.TranslateText("guid:"), cmd.serviceInstance.GUID},
		{cmd.UI.TranslateText("type:"), string(cmd.serviceInstance.Type)},
		{cmd.UI.TranslateText("tags:"), cmd.serviceInstance.Tags.String()},
		{cmd.UI.TranslateText("route service url:"), cmd.serviceInstance.RouteServiceURL.String()},
		{cmd.UI.TranslateText("syslog drain url:"), cmd.serviceInstance.SyslogDrainURL.String()},
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)
	return nil
}

func (cmd ServiceCommand) displayInto() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Showing info of service {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.serviceInstance.Name,
			"OrgName":             cmd.Config.TargetedOrganization().Name,
			"SpaceName":           cmd.Config.TargetedSpace().Name,
			"Username":            user.Name,
		},
	)
	cmd.UI.DisplayNewline()
	return nil
}
