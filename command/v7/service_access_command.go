package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . ServiceAccessActor

type ServiceAccessActor interface {
	GetServiceAccess(broker, service, organization string) ([]v7action.ServicePlanAccess, v7action.Warnings, error)
}

type ServiceAccessCommand struct {
	Broker          string      `short:"b" description:"Access for plans of a particular broker"`
	ServiceOffering string      `short:"e" description:"Access for service name of a particular service offering"`
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
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd ServiceAccessCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	servicePlanAccess, warnings, err := cmd.Actor.GetServiceAccess(cmd.Broker, cmd.ServiceOffering, cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(servicePlanAccess) == 0 {
		cmd.UI.DisplayText("No service access found.")
		return nil
	}

	tableHeaders := []string{"service", "plan", "access", "orgs"}
	data := [][]string{tableHeaders}

	for index, plan := range servicePlanAccess {
		data = append(data, []string{
			plan.ServiceOfferingName,
			plan.ServicePlanName,
			accessFromVisibilityType(plan.VisibilityType),
			strings.Join(plan.VisibilityDetails, ","),
		})

		endOfList := (index + 1) == len(servicePlanAccess)

		endOfGrouping := endOfList || plan.BrokerName != servicePlanAccess[index+1].BrokerName

		if endOfGrouping {
			cmd.UI.DisplayText("broker: {{.BrokerName}}", map[string]interface{}{
				"BrokerName": plan.BrokerName,
			})
			cmd.UI.DisplayTableWithHeader("   ", data, 3)
			cmd.UI.DisplayNewline()

			data = [][]string{tableHeaders}
		}
	}

	return nil
}

func accessFromVisibilityType(visibilityType string) string {
	switch visibilityType {
	case "public":
		return "all"
	case "admin":
		return "none"
	default:
		return "limited"
	}
}

func (cmd ServiceAccessCommand) displayMessage() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	serviceAccessMessage{
		Method:          "Getting",
		Broker:          cmd.Broker,
		ServiceOffering: cmd.ServiceOffering,
		Organization:    cmd.Organization,
		User:            user.Name,
	}.displayMessage(cmd.UI)

	cmd.UI.DisplayNewline()

	return nil
}

type serviceAccessMessage struct {
	Method, Broker, ServiceOffering, Organization, User string
}

func (msg serviceAccessMessage) displayMessage(ui command.UI) {
	var resources []string

	if msg.Broker != "" {
		resources = append(resources, "broker {{.Broker}}")
	}

	if msg.ServiceOffering != "" {
		resources = append(resources, "service {{.ServiceOffering}}")
	}

	if msg.Organization != "" {
		resources = append(resources, "organization {{.Org}}")
	}

	// We don't want the method displayed with "flavor"
	template := msg.Method + " service access"

	if len(resources) != 0 {
		template += " for " + strings.Join(resources, " and ")
	}

	template += " as {{.CurrentUser}}..."

	ui.DisplayTextWithFlavor(template, map[string]interface{}{
		"Broker":          msg.Broker,
		"ServiceOffering": msg.ServiceOffering,
		"Org":             msg.Organization,
		"CurrentUser":     msg.User,
	})
}
