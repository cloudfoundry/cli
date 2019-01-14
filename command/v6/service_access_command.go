package v6

import (
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/composite"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/sorting"
)

//go:generate counterfeiter . ServiceAccessActor

type ServiceAccessActor interface {
	GetServiceBrokerSummaries(broker string, service string, organization string) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error)
}

type ServiceAccessCommand struct {
	Broker          string      `short:"b" description:"Access for plans of a particular broker"`
	Service         string      `short:"e" description:"Access for service name of a particular service offering"`
	Organization    string      `short:"o" description:"Plans accessible by a particular organization"`
	usage           interface{} `usage:"CF_NAME service-access [-b BROKER] [-e SERVICE] [-o ORG]"`
	relatedCommands interface{} `related_commands:"marketplace, disable-service-access, enable-service-access, service-brokers"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ServiceAccessActor
}

func (cmd *ServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	baseActor := v2action.NewActor(ccClient, uaaClient, config)
	cmd.Actor = &composite.ServiceBrokerSummaryCompositeActor{
		ServiceActor:    baseActor,
		BrokerActor:     baseActor,
		OrgActor:        baseActor,
		VisibilityActor: baseActor,
	}

	return nil
}

func (cmd ServiceAccessCommand) Execute(args []string) error {
	if !cmd.Config.Experimental() {
		return translatableerror.UnrefactoredCommandError{}
	}

	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	template := "Getting service access as {{.CurrentUser}}..."
	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"CurrentUser": user.Name,
	})

	summaries, warnings, err := cmd.Actor.GetServiceBrokerSummaries(cmd.Broker, cmd.Service, cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	sortBrokers(summaries)

	tableHeaders := []string{"service", "plan", "access", "orgs"}
	for _, broker := range summaries {
		cmd.UI.DisplayText("broker: {{.BrokerName}}", map[string]interface{}{
			"BrokerName": broker.Name,
		})

		data := [][]string{tableHeaders}

		for _, service := range broker.Services {
			for _, plan := range service.Plans {
				data = append(data, []string{
					service.Label,
					plan.Name,
					cmd.UI.TranslateText(formatAccess(plan)),
					strings.Join(plan.VisibleTo, ","),
				})
			}
		}

		cmd.UI.DisplayTableWithHeader("   ", data, 3)
		cmd.UI.DisplayNewline()
	}

	return nil
}

func formatAccess(plan v2action.ServicePlanSummary) string {
	if plan.Public {
		return "all"
	}

	if len(plan.VisibleTo) > 0 {
		return "limited"
	}

	return "none"
}

func sortBrokers(brokers []v2action.ServiceBrokerSummary) {
	sort.SliceStable(brokers, func(i, j int) bool {
		return sorting.LessIgnoreCase(brokers[i].Name, brokers[j].Name)
	})

	for _, broker := range brokers {
		sortServices(broker.Services)
	}
}

func sortServices(services []v2action.ServiceSummary) {
	sort.SliceStable(services, func(i, j int) bool {
		return sorting.LessIgnoreCase(services[i].Label, services[j].Label)
	})

	for _, service := range services {
		sortPlans(service.Plans)
	}
}

func sortPlans(plans []v2action.ServicePlanSummary) {
	sort.SliceStable(plans, func(i, j int) bool {
		return sorting.LessIgnoreCase(plans[i].Name, plans[j].Name)
	})

	for _, plan := range plans {
		sortOrgs(plan.VisibleTo)
	}
}

func sortOrgs(orgs []string) {
	sort.Strings(orgs)
}
