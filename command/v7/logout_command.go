package v7

type LogoutCommand struct {
	BaseCommand

	usage interface{} `usage:"CF_NAME logout"`
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
	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayOK()

	return nil
}
