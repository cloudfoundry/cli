package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableFeatureFlagCommand struct {
	BaseCommand

	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, feature-flag, feature-flags"`
}

func (cmd EnableFeatureFlagCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Enabling feature flag {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": cmd.RequiredArgs.Feature,
		"Username":    user.Name,
	})

	warnings, err := cmd.Actor.EnableFeatureFlag(cmd.RequiredArgs.Feature)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
