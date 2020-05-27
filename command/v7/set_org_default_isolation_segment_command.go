package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SetOrgDefaultIsolationSegmentCommand struct {
	BaseCommand

	RequiredArgs    flag.OrgIsolationArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME set-org-default-isolation-segment ORG_NAME SEGMENT_NAME"`
	relatedCommands interface{}           `related_commands:"org, set-space-isolation-segment"`
}

func (cmd SetOrgDefaultIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting isolation segment {{.IsolationSegmentName}} to default on org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"OrgName":              cmd.RequiredArgs.OrganizationName,
		"CurrentUser":          user.Name,
	})

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	isoSeg, warnings, err := cmd.Actor.GetIsolationSegmentByName(cmd.RequiredArgs.IsolationSegmentName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.SetOrganizationDefaultIsolationSegment(org.GUID, isoSeg.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Restart applications in this organization to relocate them to this isolation segment.")

	return nil
}
