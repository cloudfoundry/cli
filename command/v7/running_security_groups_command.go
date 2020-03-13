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

//go:generate counterfeiter . RunningSecurityGroupsActor

type RunningSecurityGroupsActor interface {
	GetGlobalRunningSecurityGroups() ([]resources.SecurityGroup, v7action.Warnings, error)
}

type RunningSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME running-security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, security-group, unbind-running-security-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       RunningSecurityGroupsActor
}

func (cmd *RunningSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
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
