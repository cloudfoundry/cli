package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/ui"
)

type OrgsCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME orgs [--labels SELECTOR]\n\nEXAMPLES:\n   CF_NAME orgs\n   CF_NAME orgs --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME orgs --labels 'env=dev,!chargeback-code,tier in (backend,worker)'"`
	relatedCommands interface{} `related_commands:"create-org, org, org-users, set-org-role"`
	Labels      string `long:"labels" description:"Selector to filter orgs by labels"`
}

func (cmd OrgsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting orgs as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	orgs, warnings, err := cmd.Actor.GetOrganizations(cmd.Labels)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		cmd.UI.DisplayText("No orgs found.")
	} else {
		cmd.displayOrgs(orgs)
	}

	return nil
}

func (cmd OrgsCommand) displayOrgs(orgs []v7action.Organization) {
	table := [][]string{{cmd.UI.TranslateText("name")}}
	for _, org := range orgs {
		table = append(table, []string{org.Name})
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
