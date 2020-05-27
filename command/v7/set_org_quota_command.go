package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetOrgQuotaCommand struct {
	BaseCommand

	RequiredArgs    flag.SetOrgQuotaArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME set-org-quota ORG QUOTA"`
	relatedCommands interface{}          `related_commands:"org-quotas, orgs"`
}

func (cmd *SetOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting quota {{.QuotaName}} to org {{.OrgName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.OrganizationQuota,
		"OrgName":   cmd.RequiredArgs.Organization,
		"UserName":  currentUser,
	})

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.ApplyOrganizationQuotaByName(cmd.RequiredArgs.OrganizationQuota, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
