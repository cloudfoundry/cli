package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . StagingSecurityGroupsActor

type StagingSecurityGroupsActor interface {
	GetGlobalStagingSecurityGroups() ([]resources.SecurityGroup, v7action.Warnings, error)
}

type StagingSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME staging-security-groups"`
	relatedCommands interface{} `related_commands:"bind-staging-security-group, security-group, unbind-staging-security-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       StagingSecurityGroupsActor
}

func (cmd *StagingSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
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
