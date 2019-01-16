package v6

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . ServiceInstancesActor

type ServiceInstancesActor interface {
	GetServiceInstancesSummaryBySpace(spaceGUID string) ([]v2action.ServiceInstanceSummary, v2action.Warnings, error)
}

type ServicesCommand struct {
	usage           interface{} `usage:"CF_NAME services"`
	relatedCommands interface{} `related_commands:"create-service, marketplace"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ServiceInstancesActor
}

func (cmd *ServicesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd ServicesCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     cmd.Config.TargetedOrganization().Name,
			"SpaceName":   cmd.Config.TargetedSpace().Name,
			"CurrentUser": user.Name,
		})
	cmd.UI.DisplayNewline()

	instanceSummaries, warnings, err := cmd.Actor.GetServiceInstancesSummaryBySpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(instanceSummaries) == 0 {
		cmd.UI.DisplayText("No services found")
		return nil
	}

	table := [][]string{{
		cmd.UI.TranslateText("name"),
		cmd.UI.TranslateText("service"),
		cmd.UI.TranslateText("plan"),
		cmd.UI.TranslateText("bound apps"),
		cmd.UI.TranslateText("last operation"),
		cmd.UI.TranslateText("broker"),
	}}

	var boundAppNames []string

	for _, summary := range instanceSummaries {
		serviceLabel := summary.Service.Label
		if summary.ServiceInstance.Type == constant.ServiceInstanceTypeUserProvidedService {
			serviceLabel = "user-provided"
		}

		boundAppNames = []string{}
		for _, boundApplication := range summary.BoundApplications {
			boundAppNames = append(boundAppNames, boundApplication.AppName)
		}

		table = append(table, []string{
			summary.Name,
			serviceLabel,
			summary.ServicePlan.Name,
			strings.Join(boundAppNames, ", "),
			fmt.Sprintf("%s %s", summary.LastOperation.Type, summary.LastOperation.State),
			summary.Service.ServiceBrokerName},
		)
	}
	cmd.UI.DisplayTableWithHeader("", table, 3)

	return nil
}
