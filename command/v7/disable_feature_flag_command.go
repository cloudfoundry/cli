package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . DisableFeatureFlagActor

type DisableFeatureFlagActor interface {
	DisableFeatureFlag(flagName string) (v7action.Warnings, error)
}

type DisableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-feature-flag FEATURE_FLAG_NAME"`
	relatedCommands interface{}  `related_commands:"enable-feature-flag, feature-flag, feature-flags"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DisableFeatureFlagActor
}

func (cmd *DisableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
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
