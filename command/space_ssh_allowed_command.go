package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type SpaceSSHAllowedCommand struct {
	BaseCommand

	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME space-ssh-allowed SPACE_NAME"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, ssh-enabled, ssh"`
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

	displayVal := shared.FlagBoolToString(enabled)

	cmd.UI.DisplayText(
		"ssh support is {{.DisplayVal}} in space '{{.SpaceName}}'.",
		map[string]interface{}{
			"SpaceName":  cmd.RequiredArgs.Space,
			"DisplayVal": displayVal,
		},
	)

	return nil
}
