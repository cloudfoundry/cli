package v7

import (
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
)

type OrgQuotasCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME org-quotas"`
	relatedCommands interface{} `related_commands:"org-quota"`
}

func (cmd OrgQuotasCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting org quotas as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	orgQuotas, warnings, err := cmd.Actor.GetOrganizationQuotas()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	var quotas []resources.Quota
	for _, orgQuota := range orgQuotas {
		quotas = append(quotas, resources.Quota(orgQuota.Quota))
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplayQuotasTable(quotas, "No organization quotas found.")

	return nil
}
