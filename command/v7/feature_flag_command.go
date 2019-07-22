package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . FeatureFlagActor

type FeatureFlagActor interface {
	GetFeatureFlagByName(featureFlagName string) (v7action.FeatureFlag, v7action.Warnings, error)
}

type FeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, enable-feature-flag, feature-flags"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       FeatureFlagActor
}

func (cmd *FeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd FeatureFlagCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
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

func (cmd FeatureFlagCommand) displayTable(featureFlag v7action.FeatureFlag) {
	var keyValueTable = [][]string{
		{"Features", "State"},
		{featureFlag.Name, cmd.flagBoolToString(featureFlag.Enabled)},
	}

	cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}

func (cmd FeatureFlagCommand) flagBoolToString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
