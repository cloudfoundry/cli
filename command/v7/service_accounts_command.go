package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/v9/util/ui"
)

type ServiceAccountsCommand struct {
	BaseCommand
	usage           interface{} `usage:"CF_NAME service-accounts"`
	relatedCommands interface{} `related_commands:"create-service-account"`
}

func (cmd ServiceAccountsCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Getting service accounts in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"Org": cmd.Config.TargetedOrganization().Name, "Space": cmd.Config.TargetedSpace().Name, "User": user.Name,
	})
	accounts, warnings, err := cmd.Actor.GetServiceAccountsInSpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayNewline()
	if len(accounts) == 0 {
		cmd.UI.DisplayText("No service accounts found.")
		return nil
	}
	table := [][]string{{cmd.UI.TranslateText("name"), cmd.UI.TranslateText("enabled"), cmd.UI.TranslateText("status"), cmd.UI.TranslateText("client id"), cmd.UI.TranslateText("certificate DNS SAN")}}
	for _, account := range accounts {
		table = append(table, []string{account.Name, strconv.FormatBool(account.Enabled), account.Status, account.ClientID, account.CertificateDNSSAN})
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	return nil
}
