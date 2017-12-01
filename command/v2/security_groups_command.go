package v2

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . SecurityGroupsActor

type SecurityGroupsActor interface {
	CloudControllerAPIVersion() string
	GetSecurityGroupsWithOrganizationSpaceAndLifecycle(includeStaging bool) ([]v2action.SecurityGroupWithOrganizationSpaceAndLifecycle, v2action.Warnings, error)
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
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd SecurityGroupsCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	includeStaging := true

	err = command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionLifecyleStagingV2)
	if err != nil {
		switch err.(type) {
		case translatableerror.MinimumAPIVersionNotMetError:
			includeStaging = false

		default:
			return err
		}
	}

	cmd.UI.DisplayTextWithFlavor("Getting security groups as {{.UserName}}...",
		map[string]interface{}{"UserName": user.Name})

	secGroupOrgSpaces, warnings, err := cmd.Actor.GetSecurityGroupsWithOrganizationSpaceAndLifecycle(includeStaging)
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
			currentGroupIndex++
			currentGroupIndexString = fmt.Sprintf("#%d", currentGroupIndex)
			currentGroupName = secGroupOrgSpace.SecurityGroup.Name
		}

		switch {
		case secGroupOrgSpace.Organization.Name == "" && secGroupOrgSpace.Space.Name == "" &&
			(secGroupOrgSpace.SecurityGroup.RunningDefault ||
				secGroupOrgSpace.SecurityGroup.StagingDefault):
			table = append(table, []string{
				currentGroupIndexString,
				secGroupOrgSpace.SecurityGroup.Name,
				cmd.UI.TranslateText("<all>"),
				cmd.UI.TranslateText("<all>"),
				string(secGroupOrgSpace.Lifecycle),
			})
		default:
			table = append(table, []string{
				currentGroupIndexString,
				secGroupOrgSpace.SecurityGroup.Name,
				secGroupOrgSpace.Organization.Name,
				secGroupOrgSpace.Space.Name,
				string(secGroupOrgSpace.Lifecycle),
			})
		}
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
