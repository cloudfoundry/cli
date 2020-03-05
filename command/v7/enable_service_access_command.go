package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . EnableServiceAccessActor

type EnableServiceAccessActor interface {
	EnableServiceAccess(offeringName, brokerName, orgName, planName string) (v7action.SkippedPlans, v7action.Warnings, error)
}

type EnableServiceAccessCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	ServiceBroker   string       `short:"b" description:"Enable access to a service from a particular service broker. Required when service name is ambiguous"`
	Organization    string       `short:"o" description:"Enable access for a specified organization"`
	ServicePlan     string       `short:"p" description:"Enable access to a specified service plan"`
	usage           interface{}  `usage:"CF_NAME enable-service-access SERVICE [-b BROKER] [-p PLAN] [-o ORG]"`
	relatedCommands interface{}  `related_commands:"disable-service-access, marketplace, service-access, service-brokers"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       EnableServiceAccessActor
	Config      command.Config
}

func (cmd *EnableServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd EnableServiceAccessCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	skipped, warnings, err := cmd.Actor.EnableServiceAccess(cmd.RequiredArgs.ServiceOffering, cmd.ServiceBroker, cmd.Organization, cmd.ServicePlan)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	displaySkippedPlans(cmd.UI, "all", skipped)

	cmd.UI.DisplayOK()

	return nil
}

func (cmd EnableServiceAccessCommand) displayMessage() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	setServiceAccessMessage{
		Operation:       "Enabling",
		ServiceOffering: cmd.RequiredArgs.ServiceOffering,
		ServicePlan:     cmd.ServicePlan,
		Organization:    cmd.Organization,
		ServiceBroker:   cmd.ServiceBroker,
		User:            user.Name,
	}.displayMessage(cmd.UI)

	cmd.UI.DisplayNewline()

	return nil
}

type setServiceAccessMessage struct {
	Operation, ServiceOffering, ServicePlan, Organization, ServiceBroker, User string
}

func (msg setServiceAccessMessage) displayMessage(ui command.UI) {
	template := msg.Operation + " access to "

	if msg.ServicePlan != "" {
		template += "plan {{.ServicePlan}} "
	} else {
		template += "all plans "
	}

	template += "of service {{.ServiceOffering}} "

	if msg.ServiceBroker != "" {
		template += "from broker {{.ServiceBroker}} "
	}

	if msg.Organization != "" {
		template += "for org {{.Organization}} "
	} else {
		template += "for all orgs "
	}

	template += "as {{.CurrentUser}}..."

	ui.DisplayTextWithFlavor(template, map[string]interface{}{
		"ServiceBroker":   msg.ServiceBroker,
		"ServiceOffering": msg.ServiceOffering,
		"ServicePlan":     msg.ServicePlan,
		"Organization":    msg.Organization,
		"CurrentUser":     msg.User,
	})
}

func displaySkippedPlans(ui command.UI, visibility string, skipped v7action.SkippedPlans) {
	for _, plan := range skipped {
		ui.DisplayTextWithFlavor(
			"Did not update plan {{.ServicePlan}} as it already has visibility {{.Visibility}}.",
			map[string]interface{}{
				"ServicePlan": plan,
				"Visibility":  visibility,
			},
		)
	}
	if len(skipped) > 0 {
		ui.DisplayNewline()
	}
}
