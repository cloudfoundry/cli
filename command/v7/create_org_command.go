package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateOrgCommand struct {
	BaseCommand

	RequiredArgs    flag.Organization `positional-args:"yes"`
	Quota           string            `short:"q" long:"quota" description:"Quota to assign to the newly created org"`
	usage           interface{}       `usage:"CF_NAME create-org ORG [-q ORG_QUOTA]"`
	relatedCommands interface{}       `related_commands:"create-space, orgs, org-quotas, set-org-role"`
}

func (cmd CreateOrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.RequiredArgs.Organization

	cmd.UI.DisplayTextWithFlavor("Creating org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"User":         user.Name,
			"Organization": orgName,
		})

	org, warnings, err := cmd.Actor.CreateOrganization(orgName)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.OrganizationNameTakenError); ok {
			cmd.UI.DisplayText(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}
	cmd.UI.DisplayOK()

	if cmd.Quota != "" {
		cmd.UI.DisplayTextWithFlavor("Setting org quota {{.Quota}} to org {{.Organization}} as {{.User}}...",
			map[string]interface{}{
				"Quota":        cmd.Quota,
				"Organization": orgName,
				"User":         user.Name,
			})
		warnings, err = cmd.Actor.ApplyOrganizationQuotaByName(cmd.Quota, org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayTextWithFlavor("Assigning role OrgManager to user {{.User}} in org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"User":         user.Name,
			"Organization": orgName,
		})
	warnings, err = cmd.Actor.CreateOrgRole(constant.OrgManagerRole, org.GUID, user.Name, user.Origin, user.IsClient)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayText(`TIP: Use 'cf target -o "{{.Organization}}"' to target new org`,
		map[string]interface{}{
			"Organization": orgName,
		})

	return nil
}
