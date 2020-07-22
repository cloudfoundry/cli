package v7

import (
	"encoding/json"
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

	cmd.serviceInstance, warnings, err = cmd.Actor.GetServiceInstanceDetails(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
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
		)
	default:
		return cmd.chain(
			cmd.displayIntro,
			cmd.displayPropertiesManaged,
			cmd.displaySharingInfo,
			cmd.displayLastOperation,
			cmd.displayParameters,
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
		{cmd.UI.TranslateText("plan:"), cmd.serviceInstance.ServicePlanName},
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
	sharedStatus := cmd.serviceInstance.SharedStatus

	cmd.UI.DisplayText("Sharing:")
	cmd.UI.DisplayNewline()

	if sharedStatus.FeatureFlagIsDisabled {
		cmd.UI.DisplayText(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`)
		cmd.UI.DisplayNewline()
	}

	if sharedStatus.OfferingDisablesSharing {
		cmd.UI.DisplayText("Service instance sharing is disabled for this service offering.")
		cmd.UI.DisplayNewline()
	}

	if sharedStatus.IsShared {
		cmd.UI.DisplayText("This service instance is currently shared.")
	} else {
		cmd.UI.DisplayText("This service instance is not currently being shared.")
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
		"No parameters are set for service instance {{.ServiceInstanceName}}...",
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
