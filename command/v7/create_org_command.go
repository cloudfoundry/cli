package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateOrgActor

type CreateOrgActor interface {
	CreateOrganization(orgName string) (v7action.Organization, v7action.Warnings, error)
	CreateOrgRole(roleType constant.RoleType, orgGUID string, userNameOrGUID string, userOrigin string, isClient bool) (v7action.Warnings, error)
	GetOrganizationByName(name string) (v7action.Organization, v7action.Warnings, error)
}

type CreateOrgCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME create-org ORG"`
	relatedCommands interface{}       `related_commands:"create-space, orgs, quotas, set-org-role"`

	UI          command.UI
	Config      command.Config
	Actor       CreateOrgActor
	SharedActor command.SharedActor
}

func (cmd *CreateOrgCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
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

	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		if _, ok := err.(ccerror.OrganizationNameTakenError); ok {
			cmd.UI.DisplayText(err.Error())
			org, warnings, err = cmd.Actor.GetOrganizationByName(orgName)
			cmd.UI.DisplayWarningsV7(warnings)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Assigning role OrgManager to user {{.User}} in org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"User":         user.Name,
			"Organization": orgName,
		})
	warnings, err = cmd.Actor.CreateOrgRole(constant.OrgManagerRole, org.GUID, user.Name, user.Origin, user.IsClient)
	cmd.UI.DisplayWarningsV7(warnings)
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
