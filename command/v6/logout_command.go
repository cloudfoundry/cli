package v6

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

type LogoutCommand struct {
	usage interface{} `usage:"CF_NAME logout"`

	UI     command.UI
	Config command.Config
	Actor  AuthActor
}

func (cmd *LogoutCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	_, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(nil, uaaClient, config)

	return nil
}

func (cmd LogoutCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Logging out {{.Username}}...",
		map[string]interface{}{
			"Username": user.Name,
		})

	// JG 8/4/2020 Intentionally Ignoring the error return of this,
	// even if we fail to revoke tokens log out should continue
	_ = cmd.Actor.Revoke()
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
