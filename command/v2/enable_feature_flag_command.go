package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-feature-flag FEATURE_NAME"`
	relatedCommands interface{}  `related_commands:"disable-feature-flag, feature-flags"`
}

func (_ EnableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ EnableFeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
