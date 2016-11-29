package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameBuildpackCommand struct {
	RequiredArgs    flag.RenameBuildpackArgs `positional-args:"yes"`
	usage           interface{}              `usage:"CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"`
	relatedCommands interface{}              `related_commands:"update-buildpack"`
}

func (_ RenameBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RenameBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
