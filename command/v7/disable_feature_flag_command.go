package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type DisableFeatureFlagCommand struct {
	command.BaseCommand

	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"enable-feature-flag, feature-flag, feature-flags"`
}

func (cmd DisableFeatureFlagCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Disabling feature flag {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": cmd.RequiredArgs.Feature,
		"Username":    user.Name,
	})

	warnings, err := cmd.Actor.DisableFeatureFlag(cmd.RequiredArgs.Feature)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
