package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SetOrgQuotaActor

type SetOrgQuotaActor interface {
	GetOrganizationByName(orgName string) (v7action.Organization, v7action.Warnings, error)
	ApplyOrganizationQuotaByName(quotaName, orgGUID string) (v7action.Warnings, error)
}

type SetOrgQuotaCommand struct {
	RequiredArgs    flag.SetOrgQuotaArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME set-org-quota ORG QUOTA"`
	relatedCommands interface{}          `related_commands:"org-quotas, orgs"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetOrgQuotaActor
}

func (cmd *SetOrgQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *SetOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting quota {{.QuotaName}} to org {{.OrgName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.OrganizationQuota,
		"OrgName":   cmd.RequiredArgs.Organization,
		"UserName":  currentUser,
	})

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.ApplyOrganizationQuotaByName(cmd.RequiredArgs.OrganizationQuota, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
