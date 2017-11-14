package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DisableSSHCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME disable-ssh APP_NAME"`
	relatedCommands interface{}  `related_commands:"disallow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (DisableSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DisableSSHCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
