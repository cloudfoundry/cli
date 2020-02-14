package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/command/flag"
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
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)
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
