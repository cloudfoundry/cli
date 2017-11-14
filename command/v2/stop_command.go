package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type StopCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME stop APP_NAME"`
	relatedCommands interface{}  `related_commands:"restart, scale, start"`
}

func (StopCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (StopCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
