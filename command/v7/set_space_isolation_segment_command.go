package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetSpaceIsolationSegmentCommand struct {
	BaseCommand

	RequiredArgs    flag.SpaceIsolationArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME set-space-isolation-segment SPACE_NAME SEGMENT_NAME"`
	relatedCommands interface{}             `related_commands:"org, reset-space-isolation-segment, restart, set-org-default-isolation-segment, space"`
}

func (cmd SetSpaceIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating isolation segment of space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"SpaceName":   cmd.RequiredArgs.SpaceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})

	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.RequiredArgs.SpaceName, cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.AssignIsolationSegmentToSpaceByNameAndSpace(cmd.RequiredArgs.IsolationSegmentName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Restart applications in this space to relocate them to this isolation segment.")

	return nil
}
