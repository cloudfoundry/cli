package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . RenameActor

type RenameActor interface {
	RenameApplicationByNameAndSpaceGUID(oldAppName, newAppName, spaceGUID string) (v7action.Application, v7action.Warnings, error)
}

type RenameCommand struct {
	RequiredArgs    flag.Rename `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME rename APP_NAME NEW_APP_NAME"`
	relatedCommands interface{} `related_commands:"apps, delete"`
	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
	Actor           RenameActor
}

func (cmd *RenameCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd RenameCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	oldName, newName := cmd.RequiredArgs.OldAppName, cmd.RequiredArgs.NewAppName
	cmd.UI.DisplayTextWithFlavor(
		"Renaming app {{.OldAppName}} to {{.NewAppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OldAppName": oldName,
			"NewAppName": newName,
			"Username":   user.Name,
			"OrgName":    cmd.Config.TargetedOrganization().Name,
			"SpaceName":  cmd.Config.TargetedSpace().Name,
		},
	)

	_, warnings, err := cmd.Actor.RenameApplicationByNameAndSpaceGUID(oldName, newName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	return nil
}
