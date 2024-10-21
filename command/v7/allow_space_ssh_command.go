package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type AllowSpaceSSHCommand struct {
	BaseCommand

	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME allow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"enable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (cmd *AllowSpaceSSHCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	targetedOrgGUID := cmd.Config.TargetedOrganization().GUID
	inputSpace := cmd.RequiredArgs.Space

	cmd.UI.DisplayTextWithFlavor("Enabling ssh support for space {{.Space}} as {{.CurrentUserName}}...", map[string]interface{}{
		"Space":           inputSpace,
		"CurrentUserName": currentUser.Name,
	})

	warnings, err := cmd.Actor.UpdateSpaceFeature(inputSpace, targetedOrgGUID, true, "ssh")

	cmd.UI.DisplayWarnings(warnings)

	if _, ok := err.(actionerror.SpaceSSHAlreadyEnabledError); ok {
		cmd.UI.DisplayText(err.Error())
		cmd.UI.DisplayOK()
		return nil
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return err
}
