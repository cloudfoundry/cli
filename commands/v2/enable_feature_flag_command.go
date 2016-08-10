package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type EnableFeatureFlagCommand struct {
	RequiredArgs flags.Feature `positional-args:"yes"`
	usage        interface{}   `usage:"CF_NAME enable-feature-flag FEATURE_NAME"`
}

func (_ EnableFeatureFlagCommand) Setup() error {
	return nil
}

func (_ EnableFeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
