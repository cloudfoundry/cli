package v6

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . ServiceActor

type ServiceActor interface {
	GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	GetServiceInstanceSummaryByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstanceSummary, v2action.Warnings, error)
	CloudControllerAPIVersion() string
}

type ServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	GUID            bool                 `long:"guid" description:"Retrieve and display the given service's guid. All other output for the service is suppressed."`
	usage           interface{}          `usage:"CF_NAME service SERVICE_INSTANCE"`
	relatedCommands interface{}          `related_commands:"bind-service, rename-service, update-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ServiceActor
}

func (cmd *ServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd ServiceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	if cmd.GUID {
		return cmd.displayServiceInstanceGUID()
	}

	return cmd.displayServiceInstanceSummary()
}

func (cmd ServiceCommand) displayServiceInstanceGUID() error {
	serviceInstance, warnings, err := cmd.Actor.GetServiceInstanceByNameAndSpace(string(cmd.RequiredArgs.ServiceInstance), cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(serviceInstance.GUID)
	return nil
}

func (cmd ServiceCommand) displayServiceInstanceSummary() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Showing info of service {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"OrgName":             cmd.Config.TargetedOrganization().Name,
		"SpaceName":           cmd.Config.TargetedSpace().Name,
		"UserName":            user.Name,
	})
	cmd.UI.DisplayNewline()

	serviceInstanceSummary, warnings, err := cmd.Actor.GetServiceInstanceSummaryByNameAndSpace(string(cmd.RequiredArgs.ServiceInstance), cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if serviceInstanceSummary.IsManaged() {
		cmd.displayManagedServiceInstanceSummary(serviceInstanceSummary)
		cmd.displayManagedServiceInstanceLastOperation(serviceInstanceSummary)
		cmd.displayBoundApplicationsIfExists(serviceInstanceSummary)

		isOutdated, err := command.CheckVersionOutdated(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionMaintenanceInfoInSummaryV2)
		if err != nil {
			return err
		}

		if !isOutdated {
			cmd.displayUpgradeInformation(serviceInstanceSummary)
		}

		return nil
	}

	cmd.displayUserProvidedServiceInstanceSummary(serviceInstanceSummary)
	cmd.displayBoundApplicationsIfExists(serviceInstanceSummary)
	return nil
}

func (cmd ServiceCommand) displayManagedServiceInstanceSummary(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	table := [][]string{{cmd.UI.TranslateText("name:"), serviceInstanceSummary.Name}}

	if serviceInstanceSummary.IsSharedFrom() {
		table = append(table, []string{
			cmd.UI.TranslateText("shared from org/space:"),
			fmt.Sprintf(
				"%s / %s",
				serviceInstanceSummary.ServiceInstanceSharedFrom.OrganizationName,
				serviceInstanceSummary.ServiceInstanceSharedFrom.SpaceName,
			),
		})
	}

	table = append(table, [][]string{
		{cmd.UI.TranslateText("service:"), serviceInstanceSummary.Service.Label},
		{cmd.UI.TranslateText("tags:"), strings.Join(serviceInstanceSummary.Tags, ", ")},
		{cmd.UI.TranslateText("plan:"), serviceInstanceSummary.ServicePlan.Name},
		{cmd.UI.TranslateText("description:"), serviceInstanceSummary.Service.Description},
		{cmd.UI.TranslateText("documentation:"), serviceInstanceSummary.Service.DocumentationURL},
		{cmd.UI.TranslateText("dashboard:"), serviceInstanceSummary.DashboardURL},
		{cmd.UI.TranslateText("service broker:"), serviceInstanceSummary.Service.ServiceBrokerName},
	}...)

	cmd.UI.DisplayKeyValueTable("", table, 3)

	if serviceInstanceSummary.IsNotShared() && serviceInstanceSummary.IsShareable() {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("This service is not currently shared.")
		return
	}

	if serviceInstanceSummary.IsSharedTo() {
		cmd.displayManagedServiceInstanceSharedWithInformation(serviceInstanceSummary)
	}
}

func (cmd ServiceCommand) displayManagedServiceInstanceSharedWithInformation(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	if !serviceInstanceSummary.ServiceInstanceSharingFeatureFlag || !serviceInstanceSummary.Service.Extra.Shareable {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText(translatableerror.ServiceInstanceNotShareableError{
			FeatureFlagEnabled:          serviceInstanceSummary.ServiceInstanceSharingFeatureFlag,
			ServiceBrokerSharingEnabled: serviceInstanceSummary.Service.Extra.Shareable,
		}.Error())
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("shared with spaces:")

	sharedTosTable := [][]string{{
		cmd.UI.TranslateText("org"),
		cmd.UI.TranslateText("space"),
		cmd.UI.TranslateText("bindings"),
	}}

	for _, sharedTo := range serviceInstanceSummary.ServiceInstanceSharedTos {
		sharedTosTable = append(sharedTosTable, []string{
			sharedTo.OrganizationName,
			sharedTo.SpaceName,
			fmt.Sprintf("%d", sharedTo.BoundAppCount),
		})
	}

	cmd.UI.DisplayTableWithHeader("", sharedTosTable, 3)
}

func (cmd ServiceCommand) displayManagedServiceInstanceLastOperation(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Showing status of last operation from service {{.ServiceInstanceName}}...", map[string]interface{}{"ServiceInstanceName": serviceInstanceSummary.Name})
	cmd.UI.DisplayNewline()
	lastOperationTable := [][]string{
		{cmd.UI.TranslateText("status:"), fmt.Sprintf("%s %s", serviceInstanceSummary.ServiceInstance.LastOperation.Type, serviceInstanceSummary.ServiceInstance.LastOperation.State)},
		{cmd.UI.TranslateText("message:"), serviceInstanceSummary.ServiceInstance.LastOperation.Description},
		{cmd.UI.TranslateText("started:"), serviceInstanceSummary.ServiceInstance.LastOperation.CreatedAt},
		{cmd.UI.TranslateText("updated:"), serviceInstanceSummary.ServiceInstance.LastOperation.UpdatedAt},
	}
	cmd.UI.DisplayKeyValueTable("", lastOperationTable, 3)
}

func (cmd ServiceCommand) displayUserProvidedServiceInstanceSummary(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	table := [][]string{
		{cmd.UI.TranslateText("name:"), serviceInstanceSummary.Name},
		{cmd.UI.TranslateText("service:"), cmd.UI.TranslateText("user-provided")},
		{cmd.UI.TranslateText("tags:"), strings.Join(serviceInstanceSummary.Tags, ", ")},
	}
	if serviceInstanceSummary.RouteServiceURL != "" {
		table = append(table, []string{cmd.UI.TranslateText("route service url:"), serviceInstanceSummary.RouteServiceURL})
	}
	cmd.UI.DisplayKeyValueTable("", table, 3)
}

func (cmd ServiceCommand) displayBoundApplicationsIfExists(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	cmd.UI.DisplayNewline()

	if len(serviceInstanceSummary.BoundApplications) == 0 {
		cmd.UI.DisplayText("There are no bound apps for this service.")
		return
	}

	cmd.UI.DisplayText("bound apps:")
	boundAppsTable := [][]string{{
		cmd.UI.TranslateText("name"),
		cmd.UI.TranslateText("binding name"),
		cmd.UI.TranslateText("status"),
		cmd.UI.TranslateText("message"),
	}}

	for _, boundApplication := range serviceInstanceSummary.BoundApplications {
		boundAppsTable = append(boundAppsTable, []string{
			boundApplication.AppName,
			boundApplication.ServiceBindingName,
			fmt.Sprintf("%s %s", boundApplication.LastOperation.Type, boundApplication.LastOperation.State),
			boundApplication.LastOperation.Description,
		})
	}

	cmd.UI.DisplayTableWithHeader("", boundAppsTable, 3)
}

func (cmd ServiceCommand) displayUpgradeInformation(serviceInstanceSummary v2action.ServiceInstanceSummary) {
	cmd.UI.DisplayNewline()

	if !serviceInstanceSummary.UpgradeSupported() {
		cmd.UI.DisplayText("Upgrades are not supported by this broker.")
		return
	}

	if !serviceInstanceSummary.UpgradeAvailable() {
		cmd.UI.DisplayText("There is no upgrade available for this service.")
		return
	}

	cmd.UI.DisplayText("Showing available upgrade details for this service...")
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("upgrade description: {{.Description}}", map[string]interface{}{
		"Description": serviceInstanceSummary.ServicePlan.MaintenanceInfo.Description,
	})
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: You can upgrade using 'cf update-service {{.InstanceName}} --upgrade'", map[string]interface{}{
		"InstanceName": serviceInstanceSummary.Name,
	})
}
