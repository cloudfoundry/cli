package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/util/ui"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
)

type ServiceAccessCommand struct {
	BaseCommand

	Broker          string      `short:"b" description:"Access for plans of a particular broker"`
	ServiceOffering string      `short:"e" description:"Access for service name of a particular service offering"`
	Organization    string      `short:"o" description:"Plans accessible by a particular organization"`
	usage           interface{} `usage:"CF_NAME service-access [-b BROKER] [-e SERVICE] [-o ORG]"`
	relatedCommands interface{} `related_commands:"marketplace, disable-service-access, enable-service-access, service-brokers"`
}

func (cmd ServiceAccessCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	servicePlanAccess, warnings, err := cmd.Actor.GetServiceAccess(cmd.ServiceOffering, cmd.Broker, cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(servicePlanAccess) == 0 {
		cmd.UI.DisplayText("No service plans found.")
		return nil
	}

	var data [][]string

	for index, plan := range servicePlanAccess {
		if len(data) == 0 {
			data = [][]string{getTableHeaders(plan)}
		}

		data = append(data, []string{
			plan.ServiceOfferingName,
			plan.ServicePlanName,
			accessFromVisibilityType(string(plan.VisibilityType)),
			strings.Join(plan.VisibilityDetails, ","),
		})

		endOfList := (index + 1) == len(servicePlanAccess)

		endOfGrouping := endOfList || plan.BrokerName != servicePlanAccess[index+1].BrokerName

		if endOfGrouping {
			cmd.UI.DisplayText("broker: {{.BrokerName}}", map[string]interface{}{
				"BrokerName": plan.BrokerName,
			})
			cmd.UI.DisplayTableWithHeader("   ", data, ui.DefaultTableSpacePadding)

			data = nil
		}

		if endOfGrouping && !endOfList {
			cmd.UI.DisplayNewline()
		}
	}

	return nil
}

func getTableHeaders(plan v7action.ServicePlanAccess) []string {
	if string(plan.VisibilityType) == "space" {
		return []string{"service", "plan", "access", "space"}
	}
	return []string{"service", "plan", "access", "orgs"}
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

	getServiceAccessMessage{
		Broker:          cmd.Broker,
		ServiceOffering: cmd.ServiceOffering,
		Organization:    cmd.Organization,
		User:            user.Name,
	}.displayMessage(cmd.UI)

	cmd.UI.DisplayNewline()

	return nil
}

type getServiceAccessMessage struct {
	Broker, ServiceOffering, Organization, User string
}

func (msg getServiceAccessMessage) displayMessage(ui command.UI) {
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

	template := "Getting service access"

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
