package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

type FeatureFlagCommand struct {
	BaseCommand

	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, enable-feature-flag, feature-flags"`
}

func (cmd FeatureFlagCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting info for feature flag {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": cmd.RequiredArgs.Feature,
		"Username":    user.Name,
	})
	cmd.UI.DisplayNewline()

	featureFlag, warnings, err := cmd.Actor.GetFeatureFlagByName(cmd.RequiredArgs.Feature)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayTable(featureFlag)
	return nil
}

func (cmd FeatureFlagCommand) displayTable(featureFlag resources.FeatureFlag) {
	var keyValueTable = [][]string{
		{"Features", "State"},
		{featureFlag.Name, shared.FlagBoolToString(featureFlag.Enabled)},
	}

	cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}
