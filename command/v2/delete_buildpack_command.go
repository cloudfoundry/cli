package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteBuildpackCommand struct {
	RequiredArgs    flag.Buildpack `positional-args:"yes"`
	Force           bool           `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}    `usage:"CF_NAME delete-buildpack BUILDPACK [-f]"`
	relatedCommands interface{}    `related_commands:"buildpacks"`
}

func (_ DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteBuildpackCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
