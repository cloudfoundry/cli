package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsetSpaceQuotaCommand struct {
	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME unset-space-quota SPACE SPACE_QUOTA"`
	relatedCommands interface{}            `related_commands:"space"`
}

func (UnsetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnsetSpaceQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
