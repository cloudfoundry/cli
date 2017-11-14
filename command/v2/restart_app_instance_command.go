package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RestartAppInstanceCommand struct {
	RequiredArgs    flag.AppInstance `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME restart-app-instance APP_NAME INDEX"`
	relatedCommands interface{}      `related_commands:"restart"`
}

func (RestartAppInstanceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RestartAppInstanceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
