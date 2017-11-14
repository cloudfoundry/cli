package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SpaceSSHAllowedCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME space-ssh-allowed SPACE_NAME"`
	relatedCommands interface{} `related_commands:"allow-space-ssh, ssh-enabled, ssh"`
}

func (SpaceSSHAllowedCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SpaceSSHAllowedCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
