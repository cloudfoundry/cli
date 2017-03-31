package v2

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
	sharedV3 "code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . SpaceActor

type SpaceActor interface {
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
	GetSpaceSummaryByOrganizationAndName(orgGUID string, spaceName string) (v2action.SpaceSummary, v2action.Warnings, error)
}

//go:generate counterfeiter . SpaceActorV3

type SpaceActorV3 interface {
	GetIsolationSegmentBySpace(spaceGUID string) (v3action.IsolationSegment, v3action.Warnings, error)
	CloudControllerAPIVersion() string
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
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, nil)

	ccClientV3, err := sharedV3.NewClients(config, ui, true)
	if err != nil {
		// special case for no v3 API installed
		if _, ok := err.(command.APINotFoundError); ok {
			return nil
		}
		return err
	}

	cmd.ActorV3 = v3action.NewActor(ccClientV3, config)

	return nil
}

func (cmd SpaceCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, false)

	if err == nil {
		if cmd.GUID {
			err = cmd.displaySpaceGUID()
		} else {
			err = cmd.displaySpaceSummary(cmd.SecurityGroupRules)
		}
	}

	return shared.HandleError(err)
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

	spaceSummary, warnings, err := cmd.Actor.GetSpaceSummaryByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.Space)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), spaceSummary.SpaceName},
		{cmd.UI.TranslateText("org:"), spaceSummary.OrgName},
		{cmd.UI.TranslateText("apps:"), strings.Join(spaceSummary.AppNames, ", ")},
		{cmd.UI.TranslateText("services:"), strings.Join(spaceSummary.ServiceInstanceNames, ", ")},
		{cmd.UI.TranslateText("space quota:"), spaceSummary.SpaceQuotaName},
		{cmd.UI.TranslateText("security groups:"), strings.Join(spaceSummary.SecurityGroupNames, ", ")},
	}

	if cmd.ActorV3 != nil {
		apiCheck := command.MinimumAPIVersionCheck(cmd.ActorV3.CloudControllerAPIVersion(), "3.11.0")
		if apiCheck == nil {
			isolationSegment, v3Warnings, err := cmd.ActorV3.GetIsolationSegmentBySpace(spaceSummary.SpaceGUID)
			cmd.UI.DisplayWarnings(v3Warnings)
			if err != nil {
				return sharedV3.HandleError(err)
			}

			table = append(table[:4], append([][]string{
				{cmd.UI.TranslateText("isolation segment:"), isolationSegment.Name},
			}, table[4:]...)...)
		}
	}

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
				currentGroupIndex += 1
				currentGroupIndexString = fmt.Sprintf("#%d", currentGroupIndex)
				currentGroupName = securityGroupRule.Name
			}

			table = append(table, []string{
				currentGroupIndexString,
				securityGroupRule.Name,
				securityGroupRule.Destination,
				securityGroupRule.Ports,
				securityGroupRule.Protocol,
				securityGroupRule.Lifecycle,
				securityGroupRule.Description,
			})
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTableWithHeader("", table, 3)
	}

	return nil
}
