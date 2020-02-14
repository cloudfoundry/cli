package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteUserCommand struct {
	BaseCommand

	RequiredArgs    flag.Username `positional-args:"yes"`
	Force           bool          `short:"f" description:"Prompt interactively for password"`
	Origin          string        `long:"origin" description:"Origin for mapping a user account to a user in an external identity provider"`
	usage           interface{}   `usage:"CF_NAME delete-user USERNAME [-f]\n   CF_NAME delete-user USERNAME [--origin ORIGIN]\n\nEXAMPLES:\n   cf delete-user jsmith                   # internal user\n   cf delete-user jsmith --origin ldap     # LDAP user"`
	relatedCommands interface{}   `related_commands:"org-users"`
}

func (cmd *DeleteUserCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if !cmd.Force {
		promptMessage := "Really delete the user {{.TargetUser}}?"
		deleteUser, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"TargetUser": cmd.RequiredArgs.Username})
		if promptErr != nil {
			return nil
		}

		if !deleteUser {
			cmd.UI.DisplayText("User '{{.TargetUser}}' has not been deleted.", map[string]interface{}{
				"TargetUser": cmd.RequiredArgs.Username,
			})
			cmd.UI.DisplayOK()
			return nil
		}
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting user {{.TargetUser}} as {{.CurrentUser}}...", map[string]interface{}{
		"TargetUser":  cmd.RequiredArgs.Username,
		"CurrentUser": currentUser,
	})

	user, err := cmd.Actor.GetUser(cmd.RequiredArgs.Username, cmd.Origin)
	if err != nil {
		// User never existed
		if _, ok := err.(actionerror.UserNotFoundError); ok {
			cmd.UI.DisplayTextWithFlavor(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	warnings, err := cmd.Actor.DeleteUser(user.GUID)

	if err != nil {
		return err
	}

	cmd.UI.DisplayWarnings(warnings)
	cmd.UI.DisplayOK()

	return nil
}
