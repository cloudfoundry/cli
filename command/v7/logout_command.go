package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/sharedaction"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
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
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Logging out {{.Username}}...",
		map[string]interface{}{
			"Username": user.Name,
		})

	cmd.Actor.RevokeAccessAndRefreshTokens() //nolint:errcheck
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
