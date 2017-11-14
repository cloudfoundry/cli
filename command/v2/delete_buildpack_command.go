package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteBuildpackCommand struct {
	RequiredArgs    flag.BuildpackName `positional-args:"yes"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}        `usage:"CF_NAME delete-buildpack BUILDPACK [-f]"`
	relatedCommands interface{}        `related_commands:"buildpacks"`
}

func (DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteBuildpackCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
