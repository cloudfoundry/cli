package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . DisableServiceAccessActor

type DisableServiceAccessActor interface {
	DisablePlanForOrg(serviceName, servicePlanName, orgName, brokerName string) (v2action.Warnings, error)
	DisableServiceForOrg(serviceName, orgName, brokerName string) (v2action.Warnings, error)
	DisablePlanForAllOrgs(serviceName, servicePlanName, brokerName string) (v2action.Warnings, error)
	DisableServiceForAllOrgs(serviceName, brokerName string) (v2action.Warnings, error)
}

type DisableServiceAccessCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	ServiceBroker   string       `short:"b" description:"Disable access to a service from a particular service broker. Required when service name is ambiguous"`
	Organization    string       `short:"o" description:"Disable access for a specified organization"`
	ServicePlan     string       `short:"p" description:"Disable access to a specified service plan"`
	usage           interface{}  `usage:"CF_NAME disable-service-access SERVICE [-b BROKER] [-p PLAN] [-o ORG]"`
	relatedCommands interface{}  `related_commands:"marketplace, service-access, service-brokers"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       DisableServiceAccessActor
	Config      command.Config
}

func (cmd *DisableServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd DisableServiceAccessCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	serviceName := cmd.RequiredArgs.Service
	servicePlanName := cmd.ServicePlan
	orgName := cmd.Organization
	serviceBrokerName := cmd.ServiceBroker
	var warnings v2action.Warnings

	cmd.UI.DisplayTextWithFlavor(disableMessages[disableServiceAccessOptions{servicePlanName != "", orgName != "", serviceBrokerName != ""}],
		map[string]interface{}{
			"ServicePlan":   servicePlanName,
			"Service":       serviceName,
			"ServiceBroker": serviceBrokerName,
			"Organization":  orgName,
			"User":          user.Name,
		})

	if servicePlanName != "" && orgName != "" {
		warnings, err = cmd.Actor.DisablePlanForOrg(serviceName, servicePlanName, orgName, serviceBrokerName)
	} else if orgName != "" {
		warnings, err = cmd.Actor.DisableServiceForOrg(serviceName, orgName, serviceBrokerName)
	} else if servicePlanName != "" {
		warnings, err = cmd.Actor.DisablePlanForAllOrgs(serviceName, servicePlanName, serviceBrokerName)
	} else {
		warnings, err = cmd.Actor.DisableServiceForAllOrgs(serviceName, serviceBrokerName)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

type disableServiceAccessOptions struct {
	Plan   bool
	Org    bool
	Broker bool
}

var disableMessages = map[disableServiceAccessOptions]string{
	{Plan: true, Org: true, Broker: false}:   "Disabling access to plan {{.ServicePlan}} of service {{.Service}} for org {{.Organization}} as {{.User}}...",
	{Plan: false, Org: true, Broker: false}:  "Disabling access to all plans of service {{.Service}} for the org {{.Organization}} as {{.User}}...",
	{Plan: true, Org: false, Broker: false}:  "Disabling access of plan {{.ServicePlan}} for service {{.Service}} as {{.User}}...",
	{Plan: false, Org: false, Broker: false}: "Disabling access to all plans of service {{.Service}} for all orgs as {{.User}}...",
	{Plan: true, Org: true, Broker: true}:    "Disabling access to plan {{.ServicePlan}} of service {{.Service}} from broker {{.ServiceBroker}} for org {{.Organization}} as {{.User}}...",
	{Plan: false, Org: true, Broker: true}:   "Disabling access to all plans of service {{.Service}} from broker {{.ServiceBroker}} for the org {{.Organization}} as {{.User}}...",
	{Plan: true, Org: false, Broker: true}:   "Disabling access to plan {{.ServicePlan}} of service {{.Service}} from broker {{.ServiceBroker}} for all orgs as {{.User}}...",
	{Plan: false, Org: false, Broker: true}:  "Disabling access to all plans of service {{.Service}} from broker {{.ServiceBroker}} for all orgs as {{.User}}...",
}
