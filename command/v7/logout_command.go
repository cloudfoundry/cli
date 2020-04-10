package v7

import (
	"code.cloudfoundry.org/cli/command"
)

type LogoutCommand struct {
	BaseCommand

	usage interface{} `usage:"CF_NAME logout"`
}

func (cmd *LogoutCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}
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
