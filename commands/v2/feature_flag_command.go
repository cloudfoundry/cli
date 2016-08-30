package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type FeatureFlagCommand struct {
	RequiredArgs    flags.Feature `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME feature-flag FEATURE_NAME"`
	relatedCommands interface{}   `related_commands:"disable-feature-flag, enable-feature-flag, feature-flags"`
}

func (_ FeatureFlagCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ FeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
