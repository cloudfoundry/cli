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
	cmd.UI.DisplayText("TIP: Restart applications in spaces without assigned isolation segments to move them to the platform default.")

	return nil
}
