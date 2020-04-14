package v7

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type PasswdCommand struct {
	BaseCommand

	usage interface{} `usage:"CF_NAME passwd"`
}

func (cmd PasswdCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	currentPassword, err := cmd.UI.DisplayPasswordPrompt("Current password")
	if err != nil {
		return err
	}

	newPassword, err := cmd.UI.DisplayPasswordPrompt("New password")
	if err != nil {
		return err
	}

	verifyPassword, err := cmd.UI.DisplayPasswordPrompt("Verify password")
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	cmd.UI.DisplayTextWithFlavor("Changing password for user {{.Username}}...", map[string]interface{}{
		"Username": currentUser.Name,
	})

	cmd.UI.DisplayNewline()

	if newPassword != verifyPassword {
		return translatableerror.PasswordVerificationFailedError{}
	}

	err = cmd.Actor.UpdateUserPassword(currentUser.GUID, currentPassword, newPassword)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	cmd.Config.UnsetUserInformation()
	cmd.UI.DisplayText("Please log in again.")

	return nil
}
