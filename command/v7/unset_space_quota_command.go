package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsetSpaceQuotaCommand struct {
	command.BaseCommand

	RequiredArgs    flag.UnsetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}              `usage:"CF_NAME unset-space-quota SPACE SPACE_QUOTA"`
	relatedCommands interface{}              `related_commands:"space"`
}

func (cmd *UnsetSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	targetedOrgGUID := cmd.Config.TargetedOrganization().GUID

	cmd.UI.DisplayTextWithFlavor("Unassigning space quota {{.QuotaName}} from space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.SpaceQuota,
		"SpaceName": cmd.RequiredArgs.Space,
		"UserName":  currentUser,
	})

	warnings, err := cmd.Actor.UnsetSpaceQuota(cmd.RequiredArgs.SpaceQuota, cmd.RequiredArgs.Space, targetedOrgGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
