package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SpaceSSHAllowedActor

type SpaceSSHAllowedActor interface {
	GetSpaceFeature(spaceName string, orgGUID string, feature string) (bool, v7action.Warnings, error)
}

type SpaceSSHAllowedCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME space-ssh-allowed SPACE_NAME"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, ssh-enabled, ssh"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SpaceSSHAllowedActor
}

func (cmd *SpaceSSHAllowedCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd SpaceSSHAllowedCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	enabled, warnings, err := cmd.Actor.GetSpaceFeature(cmd.RequiredArgs.Space, cmd.Config.TargetedOrganization().GUID, "ssh")
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	displayVal := "disabled"
	if enabled {
		displayVal = "enabled"
	}
	cmd.UI.DisplayText(
		"ssh support is {{.DisplayVal}} in space '{{.SpaceName}}'.",
		map[string]interface{}{
			"SpaceName":  cmd.RequiredArgs.Space,
			"DisplayVal": displayVal,
		},
	)

	return nil
}
