package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type FeatureFlagsCommand struct {
	usage interface{} `usage:"CF_NAME feature-flags"`
}

func (_ FeatureFlagsCommand) Setup() error {
	return nil
}

func (_ FeatureFlagsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
