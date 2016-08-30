package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameBuildpackCommand struct {
	RequiredArgs    flags.RenameBuildpackArgs `positional-args:"yes"`
	usage           interface{}               `usage:"CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"`
	relatedCommands interface{}               `related_commands:"update-buildpack"`
}

func (_ RenameBuildpackCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RenameBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
