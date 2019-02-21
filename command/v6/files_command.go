package v6

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type FilesCommand struct {
	RequiredArgs    flag.FilesArgs `positional-args:"yes"`
	Instance        int            `short:"i" description:"Instance"`
	usage           interface{}    `usage:"CF_NAME files APP_NAME [PATH] [-i INSTANCE]\n\nTIP:\n   To list and inspect files of an app running on the Diego backend, use 'CF_NAME ssh'"`
	relatedCommands interface{}    `related_commands:"ssh"`
	UI              command.UI
}

func (cmd *FilesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	return nil
}

func (cmd *FilesCommand) Execute(args []string) error {
	cmd.UI.DisplayFileDeprecationWarning()
	return translatableerror.UnrefactoredCommandError{}
}
