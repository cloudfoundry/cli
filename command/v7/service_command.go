package v7

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

type ServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	ShowGUID        bool                 `long:"guid" description:"Retrieve and display the given service's guid. All other output for the service is suppressed."`
	usage           interface{}          `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"bind-service, rename-service, update-service"`

	serviceInstance v7action.ServiceInstanceDetails
}

func (cmd ServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	var (
		warnings v7action.Warnings
		err      error
	)

	cmd.serviceInstance, warnings, err = cmd.Actor.GetServiceInstanceDetails(string(cmd.RequiredArgs.ServiceInstance), cmd.Config.TargetedSpace().GUID, false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	switch {
	case cmd.ShowGUID:
		return cmd.displayGUID()
	case cmd.serviceInstance.Type == resources.UserProvidedServiceInstance:
		return cmd.chain(
			cmd.displayIntro,
			cmd.displayPropertiesUserProvided,
			cmd.displayBoundApps,
		)
	default:
		return cmd.chain(
			cmd.displayIntro,
			cmd.displayPropertiesManaged,
			cmd.displayLastOperation,
			cmd.displayBoundApps,
			cmd.displayParameters,
			cmd.displaySharingInfo,
			cmd.displayUpgrades,
		)
	}
}

func (cmd ServiceCommand) displayGUID() error {
	cmd.UI.DisplayText(cmd.serviceInstance.GUID)
	return nil
}

func (cmd ServiceCommand) displayPropertiesUserProvided() error {
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

func (cmd ServiceCommand) displayPropertiesManaged() error {
	table := [][]string{
		{cmd.UI.TranslateText("name:"), cmd.serviceInstance.Name},
		{cmd.UI.TranslateText("guid:"), cmd.serviceInstance.GUID},
		{cmd.UI.TranslateText("type:"), string(cmd.serviceInstance.Type)},
		{cmd.UI.TranslateText("broker:"), cmd.serviceInstance.ServiceBrokerName},
		{cmd.UI.TranslateText("offering:"), cmd.serviceInstance.ServiceOffering.Name},
		{cmd.UI.TranslateText("plan:"), cmd.serviceInstance.ServicePlan.Name},
		{cmd.UI.TranslateText("tags:"), cmd.serviceInstance.Tags.String()},
		{cmd.UI.TranslateText("offering tags:"), cmd.serviceInstance.ServiceOffering.Tags.String()},
		{cmd.UI.TranslateText("description:"), cmd.serviceInstance.ServiceOffering.Description},
		{cmd.UI.TranslateText("documentation:"), cmd.serviceInstance.ServiceOffering.DocumentationURL},
		{cmd.UI.TranslateText("dashboard url:"), cmd.serviceInstance.DashboardURL.String()},
	}
	cmd.UI.DisplayKeyValueTable("", table, 3)

	return nil
}

func (cmd ServiceCommand) displaySharingInfo() error {
	cmd.UI.DisplayText("Sharing:")
	cmd.UI.DisplayNewline()

	sharedStatus := cmd.serviceInstance.SharedStatus

	if sharedStatus.IsSharedFromOriginalSpace {
		cmd.UI.DisplayText("This service instance is shared from space {{.Space}} of org {{.Org}}.", map[string]interface{}{
			"Space": cmd.serviceInstance.SpaceName,
			"Org":   cmd.serviceInstance.OrganizationName,
		})
		cmd.UI.DisplayNewline()
		return nil
	}

	if sharedStatus.IsSharedToOtherSpaces {
		cmd.UI.DisplayText("This service instance is currently shared.")
	} else {
		cmd.UI.DisplayText("This service instance is not currently being shared.")
	}

	if sharedStatus.FeatureFlagIsDisabled {
		cmd.UI.DisplayText(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`)
		cmd.UI.DisplayNewline()
	}

	if sharedStatus.OfferingDisablesSharing {
		cmd.UI.DisplayText("Service instance sharing is disabled for this service offering.")
		cmd.UI.DisplayNewline()
	}

	return nil
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

func (cmd ServiceCommand) displayParameters() error {
	switch {
	case cmd.serviceInstance.Parameters.MissingReason != "":
		cmd.displayParametersMissingReason()
	case len(cmd.serviceInstance.Parameters.Value.Value) > 0:
		cmd.displayParametersData()
	default:
		cmd.displayParametersEmpty()
	}

	return nil
}

func (cmd ServiceCommand) displayParametersEmpty() {
	cmd.UI.DisplayTextWithFlavor(
		"No parameters are set for service instance {{.ServiceInstanceName}}.",
		map[string]interface{}{
			"ServiceInstanceName": cmd.serviceInstance.Name,
		},
	)
}

func (cmd ServiceCommand) displayParametersData() {
	cmd.UI.DisplayTextWithFlavor(
		"Showing parameters for service instance {{.ServiceInstanceName}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.serviceInstance.Name,
		},
	)
	cmd.UI.DisplayNewline()

	data, err := json.Marshal(cmd.serviceInstance.Parameters.Value)
	if err != nil {
		panic(err)
	}

	cmd.UI.DisplayText(string(data))
}

func (cmd ServiceCommand) displayParametersMissingReason() {
	cmd.UI.DisplayText(
		"Unable to show parameters: {{.Reason}}",
		map[string]interface{}{
			"Reason": cmd.serviceInstance.Parameters.MissingReason,
		},
	)
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

	return nil
}

func (cmd ServiceCommand) displayUpgrades() error {
	cmd.UI.DisplayText("Upgrading:")

	switch cmd.serviceInstance.UpgradeStatus.State {
	case v7action.ServiceInstanceUpgradeAvailable:
		cmd.UI.DisplayText("Showing available upgrade details for this service...")
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Upgrade description: {{.Description}}", map[string]interface{}{
			"Description": cmd.serviceInstance.UpgradeStatus.Description,
		})
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("TIP: You can upgrade using 'cf upgrade-service {{.InstanceName}}'", map[string]interface{}{
			"InstanceName": cmd.serviceInstance.Name,
		})
	case v7action.ServiceInstanceUpgradeNotAvailable:
		cmd.UI.DisplayText("There is no upgrade available for this service.")
	default:
		cmd.UI.DisplayText("Upgrades are not supported by this broker.")
	}

	cmd.UI.DisplayNewline()
	return nil
}

func (cmd ServiceCommand) displayBoundApps() error {
	cmd.UI.DisplayText("Bound apps:")

	if len(cmd.serviceInstance.BoundApps) == 0 {
		cmd.UI.DisplayText("There are no bound apps for this service instance.")
		return nil
	}

	table := [][]string{{"name", "binding name", "status", "message"}}
	for _, app := range cmd.serviceInstance.BoundApps {
		table = append(table, []string{
			app.AppName,
			app.Name,
			fmt.Sprintf("%s %s", app.LastOperation.Type, app.LastOperation.State),
			app.LastOperation.Description,
		})
	}

	cmd.UI.DisplayTableWithHeader("   ", table, ui.DefaultTableSpacePadding)
	return nil
}

func (cmd ServiceCommand) chain(steps ...func() error) error {
	for i, step := range steps {
		if err := step(); err != nil {
			return err
		}

		if i < len(steps) {
			cmd.UI.DisplayNewline()
		}
	}

	return nil
}
