package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type AppsCommand struct {
	usage           interface{} `usage:"CF_NAME apps"`
	relatedCommands interface{} `related_commands:"events, logs, map-route, push, scale, start, stop, restart"`
}

func (AppsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (AppsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
