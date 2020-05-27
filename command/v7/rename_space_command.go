package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameSpaceCommand struct {
	BaseCommand

	RequiredArgs    flag.RenameSpace `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME rename-space SPACE NEW_SPACE_NAME"`
	relatedCommands interface{}      `related_commands:"space, spaces, space-quotas, space-users, target"`
}

func (cmd RenameSpaceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor(
		"Renaming space {{.OldSpaceName}} to {{.NewSpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OldSpaceName": cmd.RequiredArgs.OldSpaceName,
			"NewSpaceName": cmd.RequiredArgs.NewSpaceName,
			"Username":     user.Name,
		},
	)

	space, warnings, err := cmd.Actor.RenameSpaceByNameAndOrganizationGUID(
		cmd.RequiredArgs.OldSpaceName,
		cmd.RequiredArgs.NewSpaceName,
		cmd.Config.TargetedOrganization().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if space.GUID == cmd.Config.TargetedSpace().GUID {
		cmd.Config.V7SetSpaceInformation(space.GUID, space.Name)
	}
	cmd.UI.DisplayOK()

	return nil
}
