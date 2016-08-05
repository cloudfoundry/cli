package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type FeatureFlagCommand struct {
	RequiredArgs flags.Feature `positional-args:"yes"`
}

func (_ FeatureFlagCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
