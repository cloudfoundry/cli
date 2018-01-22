package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteBuildpackCommand struct {
	RequiredArgs    flag.BuildpackName `positional-args:"yes"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	Stack           string             `short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage           interface{}        `usage:"CF_NAME delete-buildpack BUILDPACK [-f] [-s STACK]"`
	relatedCommands interface{}        `related_commands:"buildpacks"`
}

func (DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteBuildpackCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
