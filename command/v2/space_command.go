package v2

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	sharedV3 "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . SpaceActor

type SpaceActor interface {
	CloudControllerAPIVersion() string
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
	GetSpaceSummaryByOrganizationAndName(orgGUID string, spaceName string, includeStagingSecurityGroupsRules bool) (v2action.SpaceSummary, v2action.Warnings, error)
}

//go:generate counterfeiter . SpaceActorV3

type SpaceActorV3 interface {
	CloudControllerAPIVersion() string
	GetEffectiveIsolationSegmentBySpace(spaceGUID string, orgDefaultIsolationSegmentGUID string) (v3action.IsolationSegment, v3action.Warnings, error)
}

type SpaceCommand struct {
	RequiredArgs       flag.Space  `positional-args:"yes"`
	GUID               bool        `long:"guid" description:"Retrieve and display the given space's guid.  All other output for the space is suppressed."`
	SecurityGroupRules bool        `long:"security-group-rules" description:"Retrieve the rules for all the security groups associated with the space."`
	usage              interface{} `usage:"CF_NAME space SPACE [--guid] [--security-group-rules]"`
	relatedCommands    interface{} `related_commands:"set-space-isolation-segment, space-quota, space-users"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SpaceActor
	ActorV3     SpaceActorV3
}

func (cmd *SpaceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	ccClientV3, _, err := sharedV3.NewClients(config, ui, true)
	if err != nil {
		if _, ok := err.(translatableerror.V3APIDoesNotExistError); !ok {
			return err
		}
	} else {
		cmd.ActorV3 = v3action.NewActor(ccClientV3, config, nil, nil)
	}

	return nil
}

func (cmd SpaceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)

	if err == nil {
		if cmd.GUID {
			err = cmd.displaySpaceGUID()
		} else {
			err = cmd.displaySpaceSummary(cmd.SecurityGroupRules)
		}
	}

	return err
}

func (cmd SpaceCommand) displaySpaceGUID() error {
	org, warnings, err := cmd.Actor.GetSpaceByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.Space)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(org.GUID)

	return nil
}

func (cmd SpaceCommand) displaySpaceSummary(displaySecurityGroupRules bool) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting info for space {{.TargetSpace}} in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"TargetSpace": cmd.RequiredArgs.Space,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	err = command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionLifecyleStagingV2)
	includeStagingSecurityGroupsRules := err == nil

	spaceSummary, warnings, err := cmd.Actor.GetSpaceSummaryByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.Space, includeStagingSecurityGroupsRules)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), spaceSummary.Name},
		{cmd.UI.TranslateText("org:"), spaceSummary.OrgName},
		{cmd.UI.TranslateText("apps:"), strings.Join(spaceSummary.AppNames, ", ")},
		{cmd.UI.TranslateText("services:"), strings.Join(spaceSummary.ServiceInstanceNames, ", ")},
	}

	isolationSegmentRow, err := cmd.isolationSegmentRow(spaceSummary)
	if err != nil {
		return err
	}
	if isolationSegmentRow != nil {
		table = append(table, isolationSegmentRow)
	}

	table = append(table,
		[]string{cmd.UI.TranslateText("space quota:"), spaceSummary.SpaceQuotaName})
	table = append(table,
		[]string{cmd.UI.TranslateText("running security groups:"), strings.Join(spaceSummary.RunningSecurityGroupNames, ", ")})
	table = append(table,
		[]string{cmd.UI.TranslateText("staging security groups:"), strings.Join(spaceSummary.StagingSecurityGroupNames, ", ")})

	cmd.UI.DisplayKeyValueTable("", table, 3)

	if displaySecurityGroupRules {
		table := [][]string{
			{
				cmd.UI.TranslateText(""),
				cmd.UI.TranslateText("security group"),
				cmd.UI.TranslateText("destination"),
				cmd.UI.TranslateText("ports"),
				cmd.UI.TranslateText("protocol"),
				cmd.UI.TranslateText("lifecycle"),
				cmd.UI.TranslateText("description"),
			},
		}

		currentGroupIndex := -1
		var currentGroupName string
		for _, securityGroupRule := range spaceSummary.SecurityGroupRules {
			var currentGroupIndexString string

			if securityGroupRule.Name != currentGroupName {
				currentGroupIndex++
				currentGroupIndexString = fmt.Sprintf("#%d", currentGroupIndex)
				currentGroupName = securityGroupRule.Name
			}

			table = append(table, []string{
				currentGroupIndexString,
				securityGroupRule.Name,
				securityGroupRule.Destination,
				securityGroupRule.Ports,
				securityGroupRule.Protocol,
				string(securityGroupRule.Lifecycle),
				securityGroupRule.Description,
			})
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	}

	return nil
}

func (cmd SpaceCommand) isolationSegmentRow(spaceSummary v2action.SpaceSummary) ([]string, error) {
	if cmd.ActorV3 == nil {
		return nil, nil
	}

	apiCheck := command.MinimumAPIVersionCheck(cmd.ActorV3.CloudControllerAPIVersion(), ccversion.MinVersionIsolationSegmentV3)
	if apiCheck != nil {
		return nil, nil
	}

	isolationSegmentName := ""
	isolationSegment, v3Warnings, err := cmd.ActorV3.GetEffectiveIsolationSegmentBySpace(
		spaceSummary.GUID, spaceSummary.OrgDefaultIsolationSegmentGUID)
	cmd.UI.DisplayWarnings(v3Warnings)
	if err == nil {
		isolationSegmentName = isolationSegment.Name
	} else {
		if _, ok := err.(actionerror.NoRelationshipError); !ok {
			return nil, err
		}
	}

	return []string{cmd.UI.TranslateText("isolation segment:"), isolationSegmentName}, nil
}
