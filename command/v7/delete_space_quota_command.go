package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteSpaceQuotaCommand struct {
	command.BaseCommand

	RequiredArgs    flag.Quota  `positional-args:"yes"`
	Force           bool        `long:"force" short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-space-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"space-quotas"`
}

func (cmd DeleteSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	spaceQuotaName := cmd.RequiredArgs.Quota

	if !cmd.Force {
		promptMessage := "Really delete the space quota {{.QuotaName}} in org {{.OrgName}}?"
		confirmedDelete, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{
			"QuotaName": spaceQuotaName,
			"OrgName":   cmd.Config.TargetedOrganizationName(),
		})

		if promptErr != nil {
			return promptErr
		}

		if !confirmedDelete {
			cmd.UI.DisplayText("Space quota '{{.QuotaName}}' has not been deleted.", map[string]interface{}{"QuotaName": spaceQuotaName})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting space quota {{.QuotaName}} for org {{.Org}} as {{.User}}...",
		map[string]interface{}{
			"User":      user.Name,
			"Org":       cmd.Config.TargetedOrganizationName(),
			"QuotaName": spaceQuotaName,
		})

	warnings, err := cmd.Actor.DeleteSpaceQuotaByName(spaceQuotaName, cmd.Config.TargetedOrganization().GUID)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		switch err.(type) {
		case actionerror.SpaceQuotaNotFoundForNameError:
			cmd.UI.DisplayWarning(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
