package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
)

type OrgQuotaCommand struct {
	BaseCommand

	RequiredArgs    flag.OrganizationQuota `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME org-quota QUOTA"`
	relatedCommands interface{}            `related_commands:"org, org-quotas"`
}

func (cmd OrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	quotaName := cmd.RequiredArgs.OrganizationQuotaName

	cmd.UI.DisplayTextWithFlavor(
		"Getting org quota {{.QuotaName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": quotaName,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	orgQuota, warnings, err := cmd.Actor.GetOrganizationQuotaByName(quotaName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplaySingleQuota(resources.Quota(orgQuota.Quota))

	return nil
}
