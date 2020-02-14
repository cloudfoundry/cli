package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameCommand struct {
	BaseCommand

	RequiredArgs    flag.Rename `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME rename APP_NAME NEW_APP_NAME"`
	relatedCommands interface{} `related_commands:"apps, delete"`
}


func (cmd RenameCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	oldName, newName := cmd.RequiredArgs.OldAppName, cmd.RequiredArgs.NewAppName
	cmd.UI.DisplayTextWithFlavor(
		"Renaming app {{.OldAppName}} to {{.NewAppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OldAppName": oldName,
			"NewAppName": newName,
			"Username":   user.Name,
			"OrgName":    cmd.Config.TargetedOrganization().Name,
			"SpaceName":  cmd.Config.TargetedSpace().Name,
		},
	)

	_, warnings, err := cmd.Actor.RenameApplicationByNameAndSpaceGUID(oldName, newName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	return nil
}
