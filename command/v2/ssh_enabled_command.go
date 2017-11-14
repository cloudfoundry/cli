package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SSHEnabledCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME ssh-enabled APP_NAME"`
	relatedCommands interface{}  `related_commands:"enable-ssh, space-ssh-allowed, ssh"`
}

func (SSHEnabledCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SSHEnabledCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
