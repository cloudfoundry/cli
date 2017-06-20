package v2

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . SecurityGroupsActor

type SecurityGroupsActor interface {
	GetSecurityGroupsWithOrganizationSpaceAndLifecycle() ([]v2action.SecurityGroupWithOrganizationSpaceAndLifecycle, v2action.Warnings, error)
}

type SecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME security-groups"`
	relatedCommands interface{} `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group, security-group"`

	SharedActor command.SharedActor
	Config      command.Config
	UI          command.UI
	Actor       SecurityGroupsActor
}

func (cmd *SecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd SecurityGroupsCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Getting security groups as {{.UserName}}...",
		map[string]interface{}{"UserName": user.Name})

	secGroupOrgSpaces, warnings, err := cmd.Actor.GetSecurityGroupsWithOrganizationSpaceAndLifecycle()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	table := [][]string{
		{
			cmd.UI.TranslateText(""),
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("organization"),
			cmd.UI.TranslateText("space"),
			cmd.UI.TranslateText("lifecycle"),
		},
	}

	currentGroupIndex := -1
	var currentGroupName string
	for _, secGroupOrgSpace := range secGroupOrgSpaces {
		var currentGroupIndexString string

		if secGroupOrgSpace.SecurityGroup.Name != currentGroupName {
			currentGroupIndex += 1
			currentGroupIndexString = fmt.Sprintf("#%d", currentGroupIndex)
			currentGroupName = secGroupOrgSpace.SecurityGroup.Name
		}

		table = append(table, []string{
			currentGroupIndexString,
			secGroupOrgSpace.SecurityGroup.Name,
			secGroupOrgSpace.Organization.Name,
			secGroupOrgSpace.Space.Name,
			secGroupOrgSpace.Lifecycle,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, 3)

	return nil
}
