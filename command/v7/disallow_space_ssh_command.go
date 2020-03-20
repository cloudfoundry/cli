package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . DisallowSpaceSSHActor

type DisallowSpaceSSHActor interface {
	UpdateSpaceFeature(spaceName string, orgGUID string, enabled bool, feature string) (v7action.Warnings, error)
}

type DisallowSpaceSSHCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"disable-ssh, space-ssh-allowed, ssh, ssh-enabled"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DisallowSpaceSSHActor
}

func (cmd *DisallowSpaceSSHCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, nil, clock.NewClock())
	return nil
}

func (cmd *DisallowSpaceSSHCommand) Execute(args []string) error {

	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUserName, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	targetedOrgGUID := cmd.Config.TargetedOrganization().GUID
	inputSpace := cmd.RequiredArgs.Space

	cmd.UI.DisplayTextWithFlavor("Disabling ssh support for space {{.Space}} as {{.CurrentUserName}}...", map[string]interface{}{
		"Space":           inputSpace,
		"CurrentUserName": currentUserName,
	})

	warnings, err := cmd.Actor.UpdateSpaceFeature(inputSpace, targetedOrgGUID, false, "ssh")
	cmd.UI.DisplayWarnings(warnings)

	if _, ok := err.(actionerror.SpaceSSHAlreadyDisabledError); ok {
		cmd.UI.DisplayText(err.Error())
		cmd.UI.DisplayOK()
		return nil
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return err

}
