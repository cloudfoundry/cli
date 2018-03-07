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
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Logging out {{.Username}}...",
		map[string]interface{}{
			"Username": user.Name,
		})
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
