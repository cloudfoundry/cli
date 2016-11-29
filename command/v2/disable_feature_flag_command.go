package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DisableFeatureFlagCommand struct {
	RequiredArgs    flag.Feature `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-feature-flag FEATURE_NAME"`
	relatedCommands interface{}  `related_commands:"enable-feature-flag, feature-flags"`
}

func (_ DisableFeatureFlagCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DisableFeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
