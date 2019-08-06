package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . RenameSpaceActor

type RenameSpaceActor interface {
	RenameSpaceByNameAndOrganizationGUID(oldSpaceName, newSpaceName, orgGUID string) (v7action.Space, v7action.Warnings, error)
}

type RenameSpaceCommand struct {
	RequiredArgs    flag.RenameSpace `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME rename-space SPACE NEW_SPACE_NAME"`
	relatedCommands interface{}      `related_commands:"space, spaces, space-quotas, space-users, target"`

	Config      command.Config
	UI          command.UI
	SharedActor command.SharedActor
	Actor       RenameSpaceActor
}

func (cmd *RenameSpaceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd RenameSpaceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor(
		"Renaming space {{.OldSpaceName}} to {{.NewSpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OldSpaceName": cmd.RequiredArgs.OldSpaceName,
			"NewSpaceName": cmd.RequiredArgs.NewSpaceName,
			"Username":     user.Name,
		},
	)

	space, warnings, err := cmd.Actor.RenameSpaceByNameAndOrganizationGUID(
		cmd.RequiredArgs.OldSpaceName,
		cmd.RequiredArgs.NewSpaceName,
		cmd.Config.TargetedOrganization().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if space.GUID == cmd.Config.TargetedSpace().GUID {
		cmd.Config.V7SetSpaceInformation(space.GUID, space.Name)
	}
	cmd.UI.DisplayOK()

	return nil
}
