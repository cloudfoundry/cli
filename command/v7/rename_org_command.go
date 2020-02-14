package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameOrgCommand struct {
	BaseCommand

	RequiredArgs    flag.RenameOrgArgs `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME rename-org ORG NEW_ORG_NAME"`
	relatedCommands interface{}        `related_commands:"orgs, quotas, set-org-role"`
}

func (cmd RenameOrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor(
		"Renaming org {{.OldOrgName}} to {{.NewOrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OldOrgName": cmd.RequiredArgs.OldOrgName,
			"NewOrgName": cmd.RequiredArgs.NewOrgName,
			"Username":   user.Name,
		},
	)

	org, warnings, err := cmd.Actor.RenameOrganization(
		cmd.RequiredArgs.OldOrgName,
		cmd.RequiredArgs.NewOrgName,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if org.GUID == cmd.Config.TargetedOrganization().GUID {
		cmd.Config.SetOrganizationInformation(org.GUID, org.Name)
	}
	cmd.UI.DisplayOK()

	return nil
}
