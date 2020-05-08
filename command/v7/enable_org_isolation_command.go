package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableOrgIsolationCommand struct {
	command.BaseCommand
	RequiredArgs    flag.OrgIsolationArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME enable-org-isolation ORG_NAME SEGMENT_NAME"`
	relatedCommands interface{}           `related_commands:"create-isolation-segment, isolation-segments, set-org-default-isolation-segment, set-space-isolation-segment"`
}

func (cmd EnableOrgIsolationCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Enabling isolation segment {{.SegmentName}} for org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"OrgName":     cmd.RequiredArgs.OrganizationName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.EntitleIsolationSegmentToOrganizationByName(cmd.RequiredArgs.IsolationSegmentName, cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
