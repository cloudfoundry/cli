package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DisallowSpaceSSHCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME disallow-space-ssh SPACE_NAME"`
	relatedCommands interface{} `related_commands:"disable-ssh, space-ssh-allowed, ssh, ssh-enabled"`
}

func (DisallowSpaceSSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DisallowSpaceSSHCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
