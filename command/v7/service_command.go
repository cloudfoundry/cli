package v7

import (
	"fmt"

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

	serviceInstance v7action.ServiceInstanceWithRelationships
}

func (cmd ServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	var (
		warnings v7action.Warnings
		err      error
	)

	cmd.serviceInstance, warnings, err = cmd.Actor.GetServiceInstanceDetails(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
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
		return cmd.displayManaged()
	}
}

func (cmd ServiceCommand) displayGUID() error {
	cmd.UI.DisplayText(cmd.serviceInstance.GUID)
	return nil
}

func (cmd ServiceCommand) displayUPSI() error {
	if err := cmd.displayIntro(); err != nil {
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

func (cmd ServiceCommand) displayManaged() error {
	if err := cmd.displayIntro(); err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), cmd.serviceInstance.Name},
		{cmd.UI.TranslateText("guid:"), cmd.serviceInstance.GUID},
		{cmd.UI.TranslateText("type:"), string(cmd.serviceInstance.Type)},
		{cmd.UI.TranslateText("broker:"), cmd.serviceInstance.ServiceBrokerName},
		{cmd.UI.TranslateText("offering:"), cmd.serviceInstance.ServiceOffering.Name},
		{cmd.UI.TranslateText("plan:"), cmd.serviceInstance.ServicePlanName},
		{cmd.UI.TranslateText("tags:"), cmd.serviceInstance.Tags.String()},
		{cmd.UI.TranslateText("description:"), cmd.serviceInstance.ServiceOffering.Description},
		{cmd.UI.TranslateText("documentation:"), cmd.serviceInstance.ServiceOffering.DocumentationURL},
		{cmd.UI.TranslateText("dashboard url:"), cmd.serviceInstance.DashboardURL.String()},
	}
	cmd.UI.DisplayKeyValueTable("", table, 3)
	cmd.UI.DisplayNewline()

	if err := cmd.displaySharingInfo(); err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	return cmd.displayLastOperation()
}

func (cmd ServiceCommand) displaySharingInfo() error {
	cmd.UI.DisplayText("Sharing:")

	switch sharedStatus := cmd.serviceInstance.SharedStatus.(type) {
	case v7action.ServiceIsShared:
		cmd.UI.DisplayText("This service instance is currently shared.")
	case v7action.ServiceIsNotShared:
		cmd.displayNotSharedInfo(sharedStatus)
	}

	return nil
}

func (cmd ServiceCommand) displayNotSharedInfo(sharedStatus v7action.ServiceIsNotShared) {
	if sharedStatus.FeatureFlagIsDisabled {
		cmd.UI.DisplayText(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`)
	} else {
		cmd.UI.DisplayText("This service instance is not currently being shared.")
	}
}

func (cmd ServiceCommand) displayLastOperation() error {
	cmd.UI.DisplayTextWithFlavor(
		"Showing status of last operation from service instance {{.ServiceInstanceName}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.serviceInstance.Name,
		},
	)
	cmd.UI.DisplayNewline()

	status := fmt.Sprintf("%s %s", cmd.serviceInstance.LastOperation.Type, cmd.serviceInstance.LastOperation.State)
	table := [][]string{
		{cmd.UI.TranslateText("status:"), status},
		{cmd.UI.TranslateText("message:"), cmd.serviceInstance.LastOperation.Description},
		{cmd.UI.TranslateText("started:"), cmd.serviceInstance.LastOperation.CreatedAt},
		{cmd.UI.TranslateText("updated:"), cmd.serviceInstance.LastOperation.UpdatedAt},
	}
	cmd.UI.DisplayKeyValueTable("", table, 3)

	return nil
}

func (cmd ServiceCommand) displayIntro() error {
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
