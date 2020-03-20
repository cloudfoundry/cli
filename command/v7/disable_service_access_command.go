package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

//go:generate counterfeiter . DisableServiceAccessActor

type DisableServiceAccessActor interface {
	DisableServiceAccess(offeringName, brokerName, orgName, planName string) (v7action.SkippedPlans, v7action.Warnings, error)
}

type DisableServiceAccessCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	ServiceBroker   string       `short:"b" description:"Disable access to a service from a particular service broker. Required when service name is ambiguous"`
	Organization    string       `short:"o" description:"Disable access for a specified organization"`
	ServicePlan     string       `short:"p" description:"Disable access to a specified service plan"`
	usage           interface{}  `usage:"CF_NAME disable-service-access SERVICE [-b BROKER] [-p PLAN] [-o ORG]"`
	relatedCommands interface{}  `related_commands:"enable-service-access, marketplace, service-access, service-brokers"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       DisableServiceAccessActor
	Config      command.Config
}

func (cmd *DisableServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd DisableServiceAccessCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	skipped, warnings, err := cmd.Actor.DisableServiceAccess(cmd.RequiredArgs.ServiceOffering, cmd.ServiceBroker, cmd.Organization, cmd.ServicePlan)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	displaySkippedPlans(cmd.UI, "none", skipped)

	cmd.UI.DisplayOK()

	return nil
}

func (cmd DisableServiceAccessCommand) displayMessage() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	setServiceAccessMessage{
		Operation:       "Disabling",
		ServiceOffering: cmd.RequiredArgs.ServiceOffering,
		ServicePlan:     cmd.ServicePlan,
		Organization:    cmd.Organization,
		ServiceBroker:   cmd.ServiceBroker,
		User:            user.Name,
	}.displayMessage(cmd.UI)

	cmd.UI.DisplayNewline()

	return nil
}
