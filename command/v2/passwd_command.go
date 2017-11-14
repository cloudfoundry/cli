package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type PasswdCommand struct {
	usage interface{} `usage:"CF_NAME passwd"`
}

func (PasswdCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (PasswdCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
