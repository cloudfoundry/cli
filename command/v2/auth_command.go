package v2

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . AuthActor

type AuthActor interface {
	Authenticate(config v2action.Config, username string, password string) error
}

type AuthCommand struct {
	RequiredArgs    flag.Authentication `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME auth USERNAME PASSWORD\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\nEXAMPLES:\n   CF_NAME auth name@example.com \"my password\" (use quotes for passwords with a space)\n   CF_NAME auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)"`
	relatedCommands interface{}         `related_commands:"api, login, target"`

	UI     command.UI
	Config command.Config
	Actor  AuthActor
}

func (cmd *AuthCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd AuthCommand) Execute(args []string) error {
	err := command.WarnAPIVersionCheck(cmd.Config, cmd.UI)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"API endpoint: {{.Endpoint}}",
		map[string]interface{}{
			"Endpoint": cmd.Config.Target(),
		})
	cmd.UI.DisplayText("Authenticating...")

	err = cmd.Actor.Authenticate(cmd.Config, cmd.RequiredArgs.Username, cmd.RequiredArgs.Password)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayTextWithFlavor(
		"Use '{{.Command}}' to view or set your target org and space.",
		map[string]interface{}{
			"Command": fmt.Sprintf("%s target", cmd.Config.BinaryName()),
		})

	return nil
}
