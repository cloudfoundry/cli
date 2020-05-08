package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type SpaceQuotasCommand struct {
	command.BaseCommand

	usage           interface{} `usage:"CF_NAME space-quotas"`
	relatedCommands interface{} `related_commands:"space-quota, set-space-quota"`
}

func (cmd SpaceQuotasCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting space quotas for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":  cmd.Config.TargetedOrganizationName(),
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	orgQuotas, warnings, err := cmd.Actor.GetSpaceQuotasByOrgGUID(cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	var quotas []v7action.Quota
	for _, orgQuota := range orgQuotas {
		quotas = append(quotas, v7action.Quota(orgQuota.Quota))
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplayQuotasTable(quotas, "No space quotas found.")

	return nil
}
