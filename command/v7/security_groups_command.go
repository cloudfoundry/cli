package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SecurityGroupsActor

type SecurityGroupsActor interface {
	GetSecurityGroups() ([]v7action.SecurityGroupSummary, v7action.Warnings, error)
}

type SecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, bind-security-group, bind-staging-security-group, security-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SecurityGroupsActor
}

func (cmd *SecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
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
