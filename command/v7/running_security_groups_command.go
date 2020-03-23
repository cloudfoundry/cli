package v7

import (
	"code.cloudfoundry.org/cli/util/ui"
)

type RunningSecurityGroupsCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME running-security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, security-group, unbind-running-security-group"`
}

func (cmd RunningSecurityGroupsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting global running security groups as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	runningSecurityGroups, warnings, err := cmd.Actor.GetGlobalRunningSecurityGroups()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(runningSecurityGroups) == 0 {
		cmd.UI.DisplayText("No global running security groups found.")
		return nil
	}

	table := [][]string{{
		cmd.UI.TranslateText("name"),
	}}
	for _, runningSecurityGroup := range runningSecurityGroups {
		table = append(table, []string{
			cmd.UI.TranslateText(runningSecurityGroup.Name),
		})
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
