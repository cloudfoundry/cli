package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteSecurityGroupCommand struct {
	command.BaseCommand

	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	Force           bool               `long:"force" short:"f" description:"Force deletion without confirmation"`
	usage           interface{}        `usage:"CF_NAME delete-security-group SECURITY_GROUP [-f]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}        `related_commands:"security-groups"`
}

func (cmd *DeleteSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		promptMessage := "Really delete the security group {{.securityGroup}}?"
		deleteSecurityGroup, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"securityGroup": cmd.RequiredArgs.SecurityGroup})

		if promptErr != nil {
			return promptErr
		}

		if !deleteSecurityGroup {
			cmd.UI.DisplayText("Security group '{{.securityGroup}}' has not been deleted.", map[string]interface{}{
				"securityGroup": cmd.RequiredArgs.SecurityGroup,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting security group {{.securityGroup}} as {{.username}}...", map[string]interface{}{
		"securityGroup": cmd.RequiredArgs.SecurityGroup,
		"username":      user.Name,
	})

	warnings, err := cmd.Actor.DeleteSecurityGroup(cmd.RequiredArgs.SecurityGroup)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.SecurityGroupNotFoundError:
			cmd.UI.DisplayWarning("Security group '{{.securityGroup}}' does not exist.", map[string]interface{}{
				"securityGroup": cmd.RequiredArgs.SecurityGroup,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
