package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/v7action"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util/ui"
)

type MarketplaceCommand struct {
	BaseCommand

	ServiceOfferingName string      `short:"e" description:"Show plan details for a particular service offering"`
	ServiceBrokerName   string      `short:"b" description:"Only show details for a particular service broker"`
	NoPlans             bool        `long:"no-plans" description:"Hide plan information for service offerings"`
	usage               interface{} `usage:"CF_NAME marketplace [-e SERVICE_OFFERING] [-b SERVICE_BROKER] [--no-plans]"`
	relatedCommands     interface{} `related_commands:"create-service, services"`
}

func (cmd MarketplaceCommand) Execute(args []string) error {
	var username string

	filter, err := cmd.processFlags()
	if err != nil {
		return err
	}

	if cmd.BaseCommand.SharedActor.IsLoggedIn() {
		username, err = cmd.processLoginContext()
		if err != nil {
			return err
		}

		filter.SpaceGUID = cmd.Config.TargetedSpace().GUID
	}

	cmd.displayMessage(username)

	offerings, warnings, err := cmd.BaseCommand.Actor.Marketplace(filter)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(offerings) == 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("No service offerings found.")
		return nil
	}

	switch cmd.ServiceOfferingName {
	case "":
		return cmd.displayOfferingsTable(offerings)
	default:
		return cmd.displayPlansTable(offerings)
	}
}

func (cmd MarketplaceCommand) processFlags() (v7action.MarketplaceFilter, error) {
	if cmd.ServiceOfferingName != "" && cmd.NoPlans {
		return v7action.MarketplaceFilter{}, translatableerror.ArgumentCombinationError{Args: []string{"--no-plans", "-e"}}
	}

	return v7action.MarketplaceFilter{
		ServiceOfferingName: cmd.ServiceOfferingName,
		ServiceBrokerName:   cmd.ServiceBrokerName,
	}, nil
}

func (cmd MarketplaceCommand) processLoginContext() (string, error) {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return "", err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (cmd MarketplaceCommand) displayMessage(username string) {
	var template string

	switch cmd.ServiceOfferingName {
	case "":
		template = "Getting all service offerings from marketplace"
	default:
		template = "Getting service plan information for service offering {{.ServiceOfferingName}}"
	}

	if cmd.ServiceBrokerName != "" {
		switch cmd.ServiceOfferingName {
		case "":
			template += " for "
		default:
			template += " from "
		}
		template += "service broker {{.ServiceBrokerName}}"
	}

	if username != "" {
		template += " in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}"
	}

	cmd.UI.DisplayTextWithFlavor(template+"...", map[string]interface{}{
		"ServiceOfferingName": cmd.ServiceOfferingName,
		"ServiceBrokerName":   cmd.ServiceBrokerName,
		"OrgName":             cmd.Config.TargetedOrganization().Name,
		"SpaceName":           cmd.Config.TargetedSpace().Name,
		"Username":            username,
	})
}

func (cmd MarketplaceCommand) displayPlansTable(offerings []v7action.ServiceOfferingWithPlans) error {
	for _, o := range offerings {
		data := [][]string{{"plan", "description", "free or paid"}}
		for _, p := range o.Plans {
			data = append(data, []string{p.Name, p.Description, freeOrPaid(p.Free)})
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor("broker: {{.ServiceBrokerName}}", map[string]interface{}{
			"ServiceBrokerName": o.ServiceBrokerName,
		})
		cmd.UI.DisplayTableWithHeader("   ", data, ui.DefaultTableSpacePadding)
	}

	return nil
}

func (cmd MarketplaceCommand) displayOfferingsTable(offerings []v7action.ServiceOfferingWithPlans) error {
	var data [][]string
	if cmd.NoPlans {
		data = computeOfferingsTableWithPlanNames(offerings)
	} else {
		data = computeOfferingsTableWithoutPlanNames(offerings)
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", data, ui.DefaultTableSpacePadding)
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering.")

	return nil
}

func computeOfferingsTableWithPlanNames(offerings []v7action.ServiceOfferingWithPlans) [][]string {
	data := [][]string{{"offering", "description", "broker"}}
	for _, o := range offerings {
		data = append(data, []string{o.Name, o.Description, o.ServiceBrokerName})
	}
	return data
}

func computeOfferingsTableWithoutPlanNames(offerings []v7action.ServiceOfferingWithPlans) [][]string {
	data := [][]string{{"offering", "plans", "description", "broker"}}
	for _, o := range offerings {
		data = append(data, []string{o.Name, planNames(o.Plans), o.Description, o.ServiceBrokerName})
	}
	return data
}

func planNames(plans []ccv3.ServicePlan) string {
	var names []string
	for _, p := range plans {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

func freeOrPaid(free bool) string {
	if free {
		return "free"
	}
	return "paid"
}
