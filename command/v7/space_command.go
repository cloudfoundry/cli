package v7

import (
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

type SpaceCommand struct {
	BaseCommand

	RequiredArgs       flag.Space  `positional-args:"yes"`
	GUID               bool        `long:"guid" description:"Retrieve and display the given space's guid.  All other output for the space is suppressed."`
	SecurityGroupRules bool        `long:"security-group-rules" description:"Retrieve the rules for all the security groups associated with the space."`
	usage              interface{} `usage:"CF_NAME space SPACE [--guid] [--security-group-rules]"`
	relatedCommands    interface{} `related_commands:"set-space-isolation-segment, space-quota, space-users"`
}

func (cmd SpaceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	spaceName := cmd.RequiredArgs.Space
	targetedOrg := cmd.Config.TargetedOrganization()

	if cmd.GUID {
		return cmd.displaySpaceGUID(spaceName, targetedOrg.GUID)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting info for space {{.SpaceName}} in org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"SpaceName": spaceName,
		"OrgName":   targetedOrg.Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	spaceSummary, warnings, err := cmd.Actor.GetSpaceSummaryByNameAndOrganization(spaceName, targetedOrg.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	table := [][]string{
		{cmd.UI.TranslateText("name:"), spaceSummary.Name},
		{cmd.UI.TranslateText("org:"), spaceSummary.OrgName},
		{cmd.UI.TranslateText("apps:"), strings.Join(spaceSummary.AppNames, ", ")},
		{cmd.UI.TranslateText("services:"), strings.Join(spaceSummary.ServiceInstanceNames, ", ")},
		{cmd.UI.TranslateText("isolation segment:"), spaceSummary.IsolationSegmentName},
		{cmd.UI.TranslateText("quota:"), spaceSummary.QuotaName},
		{cmd.UI.TranslateText("running security groups:"), formatSecurityGroupNames(spaceSummary.RunningSecurityGroups)},
		{cmd.UI.TranslateText("staging security groups:"), formatSecurityGroupNames(spaceSummary.StagingSecurityGroups)},
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)

	if cmd.SecurityGroupRules {
		cmd.displaySecurityGroupRulesTable(spaceSummary)
	}

	return nil
}

func (cmd SpaceCommand) displaySpaceGUID(spaceName string, orgGUID string) error {
	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(space.GUID)
	return nil
}

func formatSecurityGroupNames(groups []resources.SecurityGroup) string {
	var names []string

	for _, group := range groups {
		names = append(names, group.Name)
	}

	return strings.Join(names, ", ")
}

func (cmd SpaceCommand) displaySecurityGroupRulesTable(spaceSummary v7action.SpaceSummary) {
	tableHeaders := []string{"security group", "destination", "ports", "protocol", "lifecycle", "description"}
	table := [][]string{tableHeaders}

	rows := collectSecurityGroupRuleRows(spaceSummary)
	if len(rows) == 0 {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("No security group rules found.")
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		groupNameA := rows[i][0]
		groupNameB := rows[j][0]
		return groupNameA < groupNameB
	})

	table = append(table, rows...)

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}

func collectSecurityGroupRuleRows(spaceSummary v7action.SpaceSummary) [][]string {
	var rows [][]string

	for _, securityGroup := range spaceSummary.RunningSecurityGroups {
		for _, rule := range securityGroup.Rules {
			rows = append(rows, []string{
				securityGroup.Name,
				rule.Destination,
				nilStringPointer(rule.Ports),
				rule.Protocol,
				"running",
				nilStringPointer(rule.Description),
			})
		}
	}

	for _, securityGroup := range spaceSummary.StagingSecurityGroups {
		for _, rule := range securityGroup.Rules {
			rows = append(rows, []string{
				securityGroup.Name,
				rule.Destination,
				nilStringPointer(rule.Ports),
				rule.Protocol,
				"staging",
				nilStringPointer(rule.Description),
			})
		}
	}

	return rows
}

func nilStringPointer(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}
