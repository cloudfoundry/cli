package v6

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . ServicesSummariesActor

type ServicesSummariesActor interface {
	GetServicesSummaries() ([]v2action.ServiceSummary, v2action.Warnings, error)
	GetServicesSummariesForSpace(spaceGUID string) ([]v2action.ServiceSummary, v2action.Warnings, error)

	GetServiceSummaryByName(serviceName string) (v2action.ServiceSummary, v2action.Warnings, error)
	GetServiceSummaryForSpaceByName(spaceGUID, serviceName string) (v2action.ServiceSummary, v2action.Warnings, error)
}

type MarketplaceCommand struct {
	ServiceName     string      `short:"s" description:"Show plan details for a particular service offering"`
	usage           interface{} `usage:"CF_NAME marketplace [-s SERVICE]"`
	relatedCommands interface{} `related_commands:"create-service, services"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       ServicesSummariesActor
	Config      command.Config
}

func (cmd *MarketplaceCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *MarketplaceCommand) Execute(args []string) error {
	loggedIn := cmd.SharedActor.IsLoggedIn()

	if loggedIn {
		if !cmd.SharedActor.CheckOrgSpaceTargeted() {
			return errors.New("Cannot list marketplace services without a targeted space")
		}

		return cmd.marketplace()
	}

	return cmd.publicMarketplace()
}

func (cmd *MarketplaceCommand) marketplace() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if cmd.ServiceName == "" {
		cmd.UI.DisplayTextWithFlavor("Getting services from marketplace in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
		serviceSummaries, warnings, err := cmd.Actor.GetServicesSummariesForSpace(cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
		cmd.displayServiceSummaries(serviceSummaries)
	} else {
		cmd.UI.DisplayTextWithFlavor("Getting service plan information for service {{.ServiceName}} as {{.Username}}...",
			map[string]interface{}{
				"ServiceName": cmd.ServiceName,
				"Username":    user.Name,
			})

		serviceSummary, warnings, err := cmd.Actor.GetServiceSummaryForSpaceByName(cmd.Config.TargetedSpace().GUID, cmd.ServiceName)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
		cmd.displayServiceSummary(serviceSummary)
	}

	return nil
}

func (cmd *MarketplaceCommand) publicMarketplace() error {
	if cmd.ServiceName == "" {
		cmd.UI.DisplayText("Getting all services from marketplace...")

		serviceSummaries, warnings, err := cmd.Actor.GetServicesSummaries()
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
		cmd.displayServiceSummaries(serviceSummaries)
	} else {
		cmd.UI.DisplayTextWithFlavor("Getting service plan information for service {{.ServiceName}}...",
			map[string]interface{}{
				"ServiceName": cmd.ServiceName,
			})

		serviceSummary, warnings, err := cmd.Actor.GetServiceSummaryByName(cmd.ServiceName)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
		cmd.displayServiceSummary(serviceSummary)
	}

	return nil
}

func (cmd *MarketplaceCommand) displayServiceSummaries(serviceSummaries []v2action.ServiceSummary) {
	if len(serviceSummaries) == 0 {
		cmd.UI.DisplayText("No service offerings found")
	} else {
		tableHeaders := []string{"service", "plans", "description"}
		table := [][]string{tableHeaders}
		for _, serviceSummary := range serviceSummaries {
			table = append(table, []string{
				serviceSummary.Label,
				planNames(serviceSummary),
				serviceSummary.Description,
			})
		}

		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	}
}

func (cmd *MarketplaceCommand) displayServiceSummary(serviceSummary v2action.ServiceSummary) {
	tableHeaders := []string{"service plan", "description", "free or paid"}
	table := [][]string{tableHeaders}
	for _, plan := range serviceSummary.Plans {
		table = append(table, []string{
			plan.Name,
			plan.Description,
			formatFreeOrPaid(plan.Free),
		})
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}

func planNames(serviceSummary v2action.ServiceSummary) string {
	names := []string{}
	for _, planSummary := range serviceSummary.Plans {
		names = append(names, planSummary.Name)
	}
	return strings.Join(names, ", ")
}

func formatFreeOrPaid(free bool) string {
	if free {
		return "free"
	}
	return "paid"
}
