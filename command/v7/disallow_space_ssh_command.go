package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DisallowSpaceSSHCommand struct {
	BaseCommand

	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"disable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (cmd *DisallowSpaceSSHCommand) Execute(args []string) error {

	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUserName, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	targetedOrgGUID := cmd.Config.TargetedOrganization().GUID
	inputSpace := cmd.RequiredArgs.Space

	cmd.UI.DisplayTextWithFlavor("Disabling ssh support for space {{.Space}} as {{.CurrentUserName}}...", map[string]interface{}{
		"Space":           inputSpace,
		"CurrentUserName": currentUserName,
	})

	warnings, err := cmd.Actor.UpdateSpaceFeature(inputSpace, targetedOrgGUID, false, "ssh")
	cmd.UI.DisplayWarnings(warnings)

	if _, ok := err.(actionerror.SpaceSSHAlreadyDisabledError); ok {
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
