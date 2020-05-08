package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

type FeatureFlagsCommand struct {
	command.BaseCommand

	usage           interface{} `usage:"CF_NAME feature-flags"`
	relatedCommands interface{} `related_commands:"disable-feature-flag, enable-feature-flag, feature-flag"`
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
	cmd.UI.DisplayWarnings(warnings)
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
			state := shared.FlagBoolToString(flag.Enabled)
			keyValueTable = append(keyValueTable, []string{flag.Name, state})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
