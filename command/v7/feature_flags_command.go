package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . FeatureFlagsActor

type FeatureFlagsActor interface {
	GetFeatureFlags() ([]v7action.FeatureFlag, v7action.Warnings, error)
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
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

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

	cmd.UI.DisplayTextWithFlavor("Getting feature flags as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	flags, warnings, err := cmd.Actor.GetFeatureFlags()
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	cmd.displayTable(flags)

	return nil
}

func (cmd FeatureFlagsCommand) displayTable(featureFlags []v7action.FeatureFlag) {
	if len(featureFlags) > 0 {
		var keyValueTable = [][]string{
			{"name", "state"},
		}
		for _, flag := range featureFlags {
			state := "disabled"
			if flag.Enabled {
				state = "enabled"
			}
			keyValueTable = append(keyValueTable, []string{flag.Name, state})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
