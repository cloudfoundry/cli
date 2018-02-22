package v2

import (
	"code.cloudfoundry.org/cli/command"
)

type LogoutCommand struct {
	usage interface{} `usage:"CF_NAME logout"`

	UI     command.UI
	Config command.Config
}

func (cmd *LogoutCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd LogoutCommand) Execute(args []string) error {
	cmd.UI.DisplayText("Logging out...")
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
