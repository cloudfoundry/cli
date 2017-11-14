package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RenameBuildpackCommand struct {
	RequiredArgs    flag.RenameBuildpackArgs `positional-args:"yes"`
	usage           interface{}              `usage:"CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"`
	relatedCommands interface{}              `related_commands:"update-buildpack"`
}

func (RenameBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameBuildpackCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
