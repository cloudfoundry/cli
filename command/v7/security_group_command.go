package v7

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/ui"
)

type SecurityGroupCommand struct {
	command.BaseCommand

	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME security-group SECURITY_GROUP"`
	relatedCommands interface{}        `related_commands:"bind-running-security-group, bind-security-group, bind-staging-security-group"`
}

func (cmd SecurityGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting info for security group {{.GroupName}} as {{.Username}}...", map[string]interface{}{
		"GroupName": cmd.RequiredArgs.SecurityGroup,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	securityGroupSummary, warnings, err := cmd.Actor.GetSecurityGroupSummary(cmd.RequiredArgs.SecurityGroup)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("name:"), securityGroupSummary.Name},
		{cmd.UI.TranslateText("rules:"), ""},
	}, 3)

	jsonRules, err := json.MarshalIndent(securityGroupSummary.Rules, "\t", "\t")
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("\t" + string(jsonRules))

	cmd.UI.DisplayNewline()

	if len(securityGroupSummary.SecurityGroupSpaces) > 0 {
		table := [][]string{{"organization", "space"}}
		for _, securityGroupSpace := range securityGroupSummary.SecurityGroupSpaces {
			table = append(table, []string{securityGroupSpace.OrgName, securityGroupSpace.SpaceName})
		}
		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	} else {
		cmd.UI.DisplayText("No spaces assigned")
	}

	return nil
}
