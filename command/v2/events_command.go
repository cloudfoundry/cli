package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type EventsCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME events APP_NAME"`
}

func (EventsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (EventsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
