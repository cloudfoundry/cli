package v7

import (
	"code.cloudfoundry.org/cli/util/ui"
)

type StagingSecurityGroupsCommand struct {
	command.BaseCommand

	usage           interface{} `usage:"CF_NAME staging-security-groups"`
	relatedCommands interface{} `related_commands:"bind-staging-security-group, security-group, unbind-staging-security-group"`
}

func (cmd StagingSecurityGroupsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting global staging security groups as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	stagingSecurityGroups, warnings, err := cmd.Actor.GetGlobalStagingSecurityGroups()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(stagingSecurityGroups) == 0 {
		cmd.UI.DisplayText("No global staging security groups found.")
		return nil
	}

	table := [][]string{{
		cmd.UI.TranslateText("name"),
	}}
	for _, stagingSecurityGroup := range stagingSecurityGroups {
		table = append(table, []string{
			cmd.UI.TranslateText(stagingSecurityGroup.Name),
		})
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
