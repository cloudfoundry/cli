package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . RenameOrganizationActor

type RenameOrganizationActor interface {
	RenameOrganization(oldOrgName, newOrgName string) (v7action.Organization, v7action.Warnings, error)
}

type RenameOrgCommand struct {
	RequiredArgs    flag.RenameOrgArgs `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME rename-org ORG NEW_ORG_NAME"`
	relatedCommands interface{}        `related_commands:"orgs, quotas, set-org-role"`

	Config      command.Config
	UI          command.UI
	SharedActor command.SharedActor
	Actor       RenameOrganizationActor
}

func (cmd *RenameOrgCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
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
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	if org.GUID == cmd.Config.TargetedOrganization().GUID {
		cmd.Config.SetOrganizationInformation(org.GUID, org.Name)
	}
	cmd.UI.DisplayOK()

	return nil
}
