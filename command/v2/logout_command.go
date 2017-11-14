package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type LogoutCommand struct {
	usage interface{} `usage:"CF_NAME logout"`
}

func (LogoutCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (LogoutCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
