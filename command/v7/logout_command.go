package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

type LogoutCommand struct {
	BaseCommand

	usage interface{} `usage:"CF_NAME logout"`
}

func (cmd *LogoutCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, routingClient, _ := shared.GetNewClientsAndConnectToCF(config, ui, "")
	cmd.cloudControllerClient = ccClient
	cmd.uaaClient = uaaClient
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, routingClient, clock.NewClock())
	return nil
}

func (cmd LogoutCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Logging out {{.Username}}...",
		map[string]interface{}{
			"Username": user.Name,
		})

	cmd.Actor.RevokeAccessAndRefreshTokens()
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
