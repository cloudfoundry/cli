package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type ResetOrgDefaultIsolationSegmentCommand struct {
	BaseCommand

	RequiredArgs    flag.ResetOrgDefaultIsolationArgs `positional-args:"yes"`
	usage           interface{}                       `usage:"CF_NAME reset-org-default-isolation-segment ORG_NAME"`
	relatedCommands interface{}                       `related_commands:"org, restart"`
}

func (cmd ResetOrgDefaultIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Resetting default isolation segment of org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"OrgName":     cmd.RequiredArgs.OrgName,
		"CurrentUser": user.Name,
	})

	organization, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.ResetOrganizationDefaultIsolationSegment(organization.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("Applications in spaces of this org that have no isolation segment assigned will be placed in the platform default isolation segment.")
	cmd.UI.DisplayText("Running applications need a restart to be moved there.")

	return nil
}
