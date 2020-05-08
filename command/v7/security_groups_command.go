package v7

import (
	"code.cloudfoundry.org/cli/util/ui"
)

type SecurityGroupsCommand struct {
	command.BaseCommand

	usage           interface{} `usage:"CF_NAME security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, bind-security-group, bind-staging-security-group, security-group"`
}

func (cmd SecurityGroupsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting security groups as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	securityGroupSummaries, warnings, err := cmd.Actor.GetSecurityGroups()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(securityGroupSummaries) == 0 {
		cmd.UI.DisplayText("No security groups found.")
		return nil
	}

	table := [][]string{{
		cmd.UI.TranslateText("name"),
		cmd.UI.TranslateText("organization"),
		cmd.UI.TranslateText("space"),
		cmd.UI.TranslateText("lifecycle"),
	}}
	for _, securityGroupSummary := range securityGroupSummaries {
		if len(securityGroupSummary.SecurityGroupSpaces) == 0 {
			table = append(table, []string{
				cmd.UI.TranslateText(securityGroupSummary.Name),
				cmd.UI.TranslateText(""),
				cmd.UI.TranslateText(""),
				cmd.UI.TranslateText(""),
			})
		}

		for _, securityGroupSpace := range securityGroupSummary.SecurityGroupSpaces {
			table = append(table, []string{
				cmd.UI.TranslateText(securityGroupSummary.Name),
				cmd.UI.TranslateText(securityGroupSpace.OrgName),
				cmd.UI.TranslateText(securityGroupSpace.SpaceName),
				cmd.UI.TranslateText(securityGroupSpace.Lifecycle),
			})
		}
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
