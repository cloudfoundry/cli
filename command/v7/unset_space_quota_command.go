package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UnsetSpaceQuotaActor

type UnsetSpaceQuotaActor interface {
	UnsetSpaceQuota(spaceQuotaName, spaceName, orgGUID string) (v7action.Warnings, error)
}

type UnsetSpaceQuotaCommand struct {
	RequiredArgs    flag.UnsetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}              `usage:"CF_NAME unset-space-quota SPACE SPACE_QUOTA"`
	relatedCommands interface{}              `related_commands:"space"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UnsetSpaceQuotaActor
}

func (cmd *UnsetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *UnsetSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	targetedOrgGUID := cmd.Config.TargetedOrganization().GUID

	cmd.UI.DisplayTextWithFlavor("Unassigning space quota {{.QuotaName}} from space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"QuotaName": cmd.RequiredArgs.SpaceQuota,
		"SpaceName": cmd.RequiredArgs.Space,
		"UserName":  currentUser,
	})

	warnings, err := cmd.Actor.UnsetSpaceQuota(cmd.RequiredArgs.SpaceQuota, cmd.RequiredArgs.Space, targetedOrgGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
