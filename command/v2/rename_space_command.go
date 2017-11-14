package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RenameSpaceCommand struct {
	RequiredArgs flag.RenameSpaceArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME rename-space SPACE NEW_SPACE"`
}

func (RenameSpaceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameSpaceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
