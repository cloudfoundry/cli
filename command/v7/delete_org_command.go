package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteOrgCommand struct {
	BaseCommand

	RequiredArgs    flag.Organization `positional-args:"yes"`
	Force           bool              `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}       `usage:"CF_NAME delete-org ORG [-f]"`
	relatedCommands interface{}       `related_commands:"create-org, orgs, quotas, set-org-role"`
}

func (cmd *DeleteOrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		promptMessage := "Really delete the org {{.OrgName}}, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers?"
		deleteOrg, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"OrgName": cmd.RequiredArgs.Organization})

		if promptErr != nil {
			return promptErr
		}

		if !deleteOrg {
			cmd.UI.DisplayText("Organization '{{.OrgName}}' has not been deleted.", map[string]interface{}{
				"OrgName": cmd.RequiredArgs.Organization,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":  cmd.RequiredArgs.Organization,
		"Username": user.Name,
	})

	warnings, err := cmd.Actor.DeleteOrganization(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.OrganizationNotFoundError:
			cmd.UI.DisplayWarning("Org '{{.OrgName}}' does not exist.", map[string]interface{}{
				"OrgName": cmd.RequiredArgs.Organization,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	if cmd.Config.TargetedOrganization().Name == cmd.RequiredArgs.Organization {
		cmd.UI.DisplayText("TIP: No org or space targeted, use '{{.CfTargetCommand}}' to target an org and space.",
			map[string]interface{}{"CfTargetCommand": cmd.Config.BinaryName() + " target -o ORG -s SPACE"})
		cmd.Config.UnsetOrganizationAndSpaceInformation()
	}

	return nil
}
