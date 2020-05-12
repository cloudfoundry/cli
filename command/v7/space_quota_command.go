package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
)

type SpaceQuotaCommand struct {
	BaseCommand

	RequiredArgs    flag.SpaceQuota `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME space-quota QUOTA"`
	relatedCommands interface{}     `related_commands:"space, space-quotas"`
}

func (cmd SpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	quotaName := cmd.RequiredArgs.SpaceQuota

	cmd.UI.DisplayTextWithFlavor(
		"Getting space quota {{.QuotaName}} for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": quotaName,
			"OrgName":   cmd.Config.TargetedOrganizationName(),
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	spaceQuota, warnings, err := cmd.Actor.GetSpaceQuotaByName(quotaName, cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplaySingleQuota(resources.Quota(spaceQuota.Quota))

	return nil
}
