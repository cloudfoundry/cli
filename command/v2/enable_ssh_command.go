package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type EnableSSHCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME enable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"allow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (EnableSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (EnableSSHCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
