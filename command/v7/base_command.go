package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

type BaseCommand struct {
	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       Actor

	cloudControllerClient *ccv3.Client
	uaaClient             *uaa.Client
}

func (cmd *BaseCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, routingClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.cloudControllerClient = ccClient
	cmd.uaaClient = uaaClient

	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, routingClient, clock.NewClock())
	return nil
}

func (cmd *BaseCommand) GetClients() (*ccv3.Client, *uaa.Client) {
	return cmd.cloudControllerClient, cmd.uaaClient
}
