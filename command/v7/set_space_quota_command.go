package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetSpaceQuotaCommand struct {
	command.BaseCommand

	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME set-space-quota SPACE QUOTA"`
	relatedCommands interface{}            `related_commands:"space, spaces, space-quotas"`
}

func (cmd *SetSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting space quota {{.QuotaName}} to space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.SpaceQuota,
		"SpaceName": cmd.RequiredArgs.Space,
		"UserName":  currentUser,
	})

	org := cmd.Config.TargetedOrganization()
	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.RequiredArgs.Space, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.ApplySpaceQuotaByName(cmd.RequiredArgs.SpaceQuota, space.GUID, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
