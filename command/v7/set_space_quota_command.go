package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SetSpaceQuotaActor

type SetSpaceQuotaActor interface {
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v7action.Space, v7action.Warnings, error)
	ApplySpaceQuotaByName(quotaName, spaceGUID string, orgGUID string) (v7action.Warnings, error)
}

type SetSpaceQuotaCommand struct {
	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME set-space-quota SPACE QUOTA"`
	relatedCommands interface{}            `related_commands:"space, spaces, space-quotas"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetSpaceQuotaActor
}

func (cmd *SetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *SetSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting space quota {{.QuotaName}} to space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.SpaceQuota,
		"SpaceName": cmd.RequiredArgs.Space,
		"UserName":  currentUser,
	})

	org := cmd.Config.TargetedOrganization()
	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.RequiredArgs.Space, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	warnings, err = cmd.Actor.ApplySpaceQuotaByName(cmd.RequiredArgs.SpaceQuota, space.GUID, org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
