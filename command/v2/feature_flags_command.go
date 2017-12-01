package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . FeatureFlagsActor

type FeatureFlagsActor interface {
	GetFeatureFlags() ([]v2action.FeatureFlag, v2action.Warnings, error)
}

type FeatureFlagsCommand struct {
	usage           interface{} `usage:"CF_NAME feature-flags"`
	relatedCommands interface{} `related_commands:"disable-feature-flag, enable-feature-flag"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       FeatureFlagsActor
}

func (cmd *FeatureFlagsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, nil, config)

	return nil
}

func (cmd FeatureFlagsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Retrieving status of all flagged features as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	featureFlags, warnings, err := cmd.Actor.GetFeatureFlags()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayFeatureFlags(featureFlags)
	return nil
}

func (cmd FeatureFlagsCommand) displayFeatureFlags(featureFlags []v2action.FeatureFlag) {
	table := [][]string{{cmd.UI.TranslateText("features"), cmd.UI.TranslateText("state")}}

	for _, flag := range featureFlags {
		table = append(table, []string{flag.Name, string(flag.State())})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
