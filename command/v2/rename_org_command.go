package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RenameOrgCommand struct {
	RequiredArgs flag.RenameOrgArgs `positional-args:"yes"`
	usage        interface{}        `usage:"CF_NAME rename-org ORG NEW_ORG"`
}

func (RenameOrgCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameOrgCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
