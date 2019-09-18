package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . CreateServiceBrokerActor

type CreateServiceBrokerActor interface {
	CreateServiceBroker(serviceBrokerName, username, password, brokerURI, spaceGUID string) (v2action.ServiceBroker, v2action.Warnings, error)
}

type CreateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                   `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	usage           interface{}            `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"`
	relatedCommands interface{}            `related_commands:"enable-service-access, service-brokers, target"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateServiceBrokerActor
}

func (cmd *CreateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CreateServiceBrokerCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	if cmd.SpaceScoped {
		if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
			return err
		}
	} else {
		if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
			return err
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	var spaceGUID string
	if cmd.SpaceScoped {
		cmd.UI.DisplayTextWithFlavor("Creating service broker {{.ServiceBrokerName}} in org {{.ServiceBrokerOrg}} / space {{.ServiceBrokerSpace}} as {{.User}}...",
			map[string]interface{}{
				"ServiceBrokerName":  cmd.RequiredArgs.ServiceBroker,
				"ServiceBrokerOrg":   cmd.Config.TargetedOrganization().Name,
				"ServiceBrokerSpace": cmd.Config.TargetedSpace().Name,
				"User":               user.Name,
			})
		spaceGUID = cmd.Config.TargetedSpace().GUID
	} else {
		cmd.UI.DisplayTextWithFlavor("Creating service broker {{.ServiceBrokerName}} as {{.User}}...",
			map[string]interface{}{
				"ServiceBrokerName": cmd.RequiredArgs.ServiceBroker,
				"User":              user.Name,
			})
	}

	_, warnings, err := cmd.Actor.CreateServiceBroker(
		cmd.RequiredArgs.ServiceBroker,
		cmd.RequiredArgs.Username,
		cmd.RequiredArgs.Password,
		cmd.RequiredArgs.URL,
		spaceGUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
