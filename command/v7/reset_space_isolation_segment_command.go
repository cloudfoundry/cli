package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type ResetSpaceIsolationSegmentCommand struct {
	BaseCommand

	RequiredArgs    flag.ResetSpaceIsolationArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME reset-space-isolation-segment SPACE_NAME"`
	relatedCommands interface{}                  `related_commands:"org, restart, space"`
}

func (cmd ResetSpaceIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Resetting isolation segment assignment of space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SpaceName":   cmd.RequiredArgs.SpaceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})

	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.RequiredArgs.SpaceName, cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	newIsolationSegmentName, warnings, err := cmd.Actor.ResetSpaceIsolationSegment(cmd.Config.TargetedOrganization().GUID, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	if newIsolationSegmentName == "" {
		cmd.UI.DisplayText("Applications in this space will be placed in the platform default isolation segment.")
	} else {
		cmd.UI.DisplayText("Applications in this space will be placed in isolation segment '{{.orgIsolationSegment}}'.", map[string]interface{}{
			"orgIsolationSegment": newIsolationSegmentName,
		})
	}
	cmd.UI.DisplayText("Running applications need a restart to be moved there.")

	return nil
}
