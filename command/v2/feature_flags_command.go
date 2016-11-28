package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type FeatureFlagsCommand struct {
	usage           interface{} `usage:"CF_NAME feature-flags"`
	relatedCommands interface{} `related_commands:"disable-feature-flag, enable-feature-flag"`
}

func (_ FeatureFlagsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ FeatureFlagsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
