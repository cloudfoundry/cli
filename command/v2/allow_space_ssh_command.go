package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type AllowSpaceSSHCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME allow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"enable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (AllowSpaceSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (AllowSpaceSSHCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
