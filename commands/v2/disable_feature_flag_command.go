package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DisableFeatureFlagCommand struct {
	RequiredArgs flags.Feature `positional-args:"yes"`
	usage        interface{}   `usage:"CF_NAME disable-feature-flag FEATURE_NAME"`
}

func (_ DisableFeatureFlagCommand) Setup() error {
	return nil
}

func (_ DisableFeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
