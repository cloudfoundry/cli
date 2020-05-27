package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteOrgQuotaCommand struct {
	BaseCommand

	RequiredArgs    flag.Quota  `positional-args:"yes"`
	Force           bool        `long:"force" short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-org-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"org-quotas"`
}

func (cmd DeleteOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgQuotaName := cmd.RequiredArgs.Quota

	if !cmd.Force {
		promptMessage := "Really delete the org quota {{.QuotaName}}?"
		confirmedDelete, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{
			"QuotaName": orgQuotaName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !confirmedDelete {
			cmd.UI.DisplayText("Organization quota '{{.QuotaName}}' has not been deleted.", map[string]interface{}{"QuotaName": orgQuotaName})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting org quota {{.QuotaName}} as {{.User}}...",
		map[string]interface{}{
			"User":      user.Name,
			"QuotaName": orgQuotaName,
		})

	warnings, err := cmd.Actor.DeleteOrganizationQuota(orgQuotaName)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		switch err.(type) {
		case actionerror.OrganizationQuotaNotFoundForNameError:
			cmd.UI.DisplayWarning(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
