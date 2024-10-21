package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

const indent = "   "

type ServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	ShowGUID        bool                 `long:"guid" description:"Retrieve and display the given service instances's guid. All other output is suppressed."`
	Params          bool                 `long:"params" description:"Retrieve and display the given service instances's parameters. All other output is suppressed."`
	usage           interface{}          `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"bind-service, rename-service, update-service"`
}

func (cmd ServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	switch {
	case cmd.ShowGUID:
		return cmd.fetchAndDisplayGUID()
	case cmd.Params:
		return cmd.fetchAndDisplayParams()
	default:
		return cmd.fetchAndDisplayDetails()
	}
}

func (cmd ServiceCommand) fetchAndDisplayGUID() error {
	serviceInstance, _, err := cmd.Actor.GetServiceInstanceByNameAndSpace(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(serviceInstance.GUID)
	return nil
}

func (cmd ServiceCommand) fetchAndDisplayParams() error {
	params, warnings, err := cmd.Actor.GetServiceInstanceParameters(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	err = cmd.UI.DisplayJSON("", params)
	if err != nil {
		return err
	}
	return nil
}

func (cmd ServiceCommand) fetchAndDisplayDetails() error {
	if err := cmd.displayIntro(); err != nil {
		return err
	}

	serviceInstanceWithDetails, warnings, err := cmd.Actor.GetServiceInstanceDetails(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
		false,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	switch {
	case serviceInstanceWithDetails.Type == resources.UserProvidedServiceInstance:
		cmd.displayPropertiesUserProvided(serviceInstanceWithDetails)
		cmd.displayLastOperation(serviceInstanceWithDetails)
		cmd.displayBoundApps(serviceInstanceWithDetails)
	default:
		cmd.displayPropertiesManaged(serviceInstanceWithDetails)
		cmd.displayLastOperation(serviceInstanceWithDetails)
		cmd.displayBoundApps(serviceInstanceWithDetails)
		cmd.displaySharingInfo(serviceInstanceWithDetails)
		cmd.displayUpgrades(serviceInstanceWithDetails)
	}

	return nil
}

func (cmd ServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Showing info of service {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
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

func (cmd ServiceCommand) displayPropertiesUserProvided(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	table := [][]string{
		{cmd.UI.TranslateText("name:"), serviceInstanceWithDetails.Name},
		{cmd.UI.TranslateText("guid:"), serviceInstanceWithDetails.GUID},
		{cmd.UI.TranslateText("type:"), string(serviceInstanceWithDetails.Type)},
		{cmd.UI.TranslateText("tags:"), serviceInstanceWithDetails.Tags.String()},
		{cmd.UI.TranslateText("route service url:"), serviceInstanceWithDetails.RouteServiceURL.String()},
		{cmd.UI.TranslateText("syslog drain url:"), serviceInstanceWithDetails.SyslogDrainURL.String()},
	}

	cmd.UI.DisplayKeyValueTable("", table, ui.DefaultTableSpacePadding)
	cmd.UI.DisplayNewline()
}

func (cmd ServiceCommand) displayPropertiesManaged(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	table := [][]string{
		{cmd.UI.TranslateText("name:"), serviceInstanceWithDetails.Name},
		{cmd.UI.TranslateText("guid:"), serviceInstanceWithDetails.GUID},
		{cmd.UI.TranslateText("type:"), string(serviceInstanceWithDetails.Type)},
		{cmd.UI.TranslateText("broker:"), serviceInstanceWithDetails.ServiceBrokerName},
		{cmd.UI.TranslateText("offering:"), serviceInstanceWithDetails.ServiceOffering.Name},
		{cmd.UI.TranslateText("plan:"), serviceInstanceWithDetails.ServicePlan.Name},
		{cmd.UI.TranslateText("tags:"), serviceInstanceWithDetails.Tags.String()},
		{cmd.UI.TranslateText("offering tags:"), serviceInstanceWithDetails.ServiceOffering.Tags.String()},
		{cmd.UI.TranslateText("description:"), serviceInstanceWithDetails.ServiceOffering.Description},
		{cmd.UI.TranslateText("documentation:"), serviceInstanceWithDetails.ServiceOffering.DocumentationURL},
		{cmd.UI.TranslateText("dashboard url:"), serviceInstanceWithDetails.DashboardURL.String()},
	}
	cmd.UI.DisplayKeyValueTable("", table, ui.DefaultTableSpacePadding)
	cmd.UI.DisplayNewline()
}

func (cmd ServiceCommand) displaySharingInfo(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	cmd.UI.DisplayText("Showing sharing info:")

	if serviceInstanceWithDetails.SharedStatus.IsSharedFromOriginalSpace {
		cmd.UI.DisplayTextWithFlavor(indent+"This service instance is shared from space {{.Space}} of org {{.Org}}.", map[string]interface{}{
			"Space": serviceInstanceWithDetails.SpaceName,
			"Org":   serviceInstanceWithDetails.OrganizationName,
		})
		cmd.UI.DisplayNewline()
		return
	}

	if serviceInstanceWithDetails.SharedStatus.IsSharedToOtherSpaces {
		cmd.UI.DisplayText(indent + "Shared with spaces:")
		cmd.displaySharedTo(serviceInstanceWithDetails)
		cmd.UI.DisplayNewline()
	} else {
		cmd.UI.DisplayText(indent + "This service instance is not currently being shared.")
		cmd.UI.DisplayNewline()
	}

	if serviceInstanceWithDetails.SharedStatus.FeatureFlagIsDisabled || serviceInstanceWithDetails.SharedStatus.OfferingDisablesSharing {
		if serviceInstanceWithDetails.SharedStatus.FeatureFlagIsDisabled {
			cmd.UI.DisplayText(indent + `The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`)
		}

		if serviceInstanceWithDetails.SharedStatus.OfferingDisablesSharing {
			cmd.UI.DisplayText(indent + "Service instance sharing is disabled for this service offering.")
		}
		cmd.UI.DisplayNewline()
	}
}

func (cmd ServiceCommand) displayLastOperation(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	cmd.UI.DisplayText("Showing status of last operation:")

	if serviceInstanceWithDetails.LastOperation == (resources.LastOperation{}) {
		cmd.UI.DisplayText(indent + "There is no last operation available for this service instance.")
	} else {
		status := fmt.Sprintf("%s %s", serviceInstanceWithDetails.LastOperation.Type, serviceInstanceWithDetails.LastOperation.State)
		table := [][]string{
			{cmd.UI.TranslateText("status:"), status},
			{cmd.UI.TranslateText("message:"), serviceInstanceWithDetails.LastOperation.Description},
			{cmd.UI.TranslateText("started:"), serviceInstanceWithDetails.LastOperation.CreatedAt},
			{cmd.UI.TranslateText("updated:"), serviceInstanceWithDetails.LastOperation.UpdatedAt},
		}
		cmd.UI.DisplayKeyValueTable(indent, table, ui.DefaultTableSpacePadding)
	}

	cmd.UI.DisplayNewline()
}

func (cmd ServiceCommand) displayUpgrades(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	cmd.UI.DisplayText("Showing upgrade status:")

	switch serviceInstanceWithDetails.UpgradeStatus.State {
	case v7action.ServiceInstanceUpgradeAvailable:
		cmd.UI.DisplayText(indent + "There is an upgrade available for this service.")
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText(indent+"Upgrade description: {{.Description}}", map[string]interface{}{
			"Description": serviceInstanceWithDetails.UpgradeStatus.Description,
		})
		cmd.UI.DisplayText(indent+"TIP: You can upgrade using 'cf upgrade-service {{.InstanceName}}'", map[string]interface{}{
			"InstanceName": serviceInstanceWithDetails.Name,
		})
	case v7action.ServiceInstanceUpgradeNotAvailable:
		cmd.UI.DisplayText(indent + "There is no upgrade available for this service.")
	default:
		cmd.UI.DisplayText(indent + "Upgrades are not supported by this broker.")
	}

	cmd.UI.DisplayNewline()
}

func (cmd ServiceCommand) displaySharedTo(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	table := [][]string{{"org", "space", "bindings"}}
	for _, usageSummaryLine := range serviceInstanceWithDetails.SharedStatus.UsageSummary {
		table = append(table, []string{usageSummaryLine.OrganizationName, usageSummaryLine.SpaceName, strconv.Itoa(usageSummaryLine.BoundAppCount)})
	}
	cmd.UI.DisplayTableWithHeader(indent, table, ui.DefaultTableSpacePadding)
}

func (cmd ServiceCommand) displayBoundApps(serviceInstanceWithDetails v7action.ServiceInstanceDetails) {
	cmd.UI.DisplayText("Showing bound apps:")

	if len(serviceInstanceWithDetails.BoundApps) == 0 {
		cmd.UI.DisplayText(indent + "There are no bound apps for this service instance.")
		cmd.UI.DisplayNewline()
		return
	}

	table := [][]string{{"name", "binding name", "status", "message"}}
	for _, app := range serviceInstanceWithDetails.BoundApps {
		table = append(table, []string{
			app.AppName,
			app.Name,
			fmt.Sprintf("%s %s", app.LastOperation.Type, app.LastOperation.State),
			app.LastOperation.Description,
		})
	}

	cmd.UI.DisplayTableWithHeader(indent, table, ui.DefaultTableSpacePadding)
	cmd.UI.DisplayNewline()
}
