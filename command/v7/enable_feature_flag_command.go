package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . EnableFeatureFlagActor

type EnableFeatureFlagActor interface {
	EnableFeatureFlag(flagName string) (v7action.Warnings, error)
}

type EnableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, feature-flag, feature-flags"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       EnableFeatureFlagActor
}

func (cmd *EnableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)

	return nil
}

func (cmd EnableFeatureFlagCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
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
	cmd.UI.DisplayTextWithFlavor("Feature flag {{.FeatureFlag}} enabled",
		map[string]interface{}{
			"FeatureFlag": cmd.RequiredArgs.Feature,
		})
	return nil
}
