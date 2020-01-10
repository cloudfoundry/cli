package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateOrgQuotaActor

type CreateOrgQuotaActor interface {
	CreateOrganizationQuota(orgQuotaName string) (v7action.OrganizationQuota, v7action.Warnings, error)
}

type CreateOrgQuotaCommand struct {
	RequiredArgs flag.OrganizationQuota `positional-args:"yes"`
	usage        interface{}            `usage:"CF_NAME create-org-quota ORG_QUOTA"`
	// relatedCommands interface{}       `related_commands:"create-space, orgs, quotas, set-org-role"`

	UI          command.UI
	Config      command.Config
	Actor       CreateOrgQuotaActor
	SharedActor command.SharedActor
}

func (cmd *CreateOrgQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CreateOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgQuotaName := cmd.RequiredArgs.OrganizationQuota

	cmd.UI.DisplayTextWithFlavor("Creating org quota {{.OrganizationQuota}} as {{.User}}...",
		map[string]interface{}{
			"User":               user.Name,
			"Organization Quota": orgQuotaName,
		})

	// org, warnings, err := cmd.Actor.CreateOrganizationQuota(orgQuotaName)

	// cmd.UI.DisplayWarningsV7(warnings)
	// if err != nil {
	// 	if _, ok := err.(ccerror.OrganizationNameTakenError); ok {
	// 		cmd.UI.DisplayText(err.Error())
	// 		cmd.UI.DisplayOK()
	// 		return nil
	// 	}
	// 	return err
	// }
	// cmd.UI.DisplayOK()

	return nil
}
