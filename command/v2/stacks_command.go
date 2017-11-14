package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type StacksCommand struct {
	usage           interface{} `usage:"CF_NAME stacks"`
	relatedCommands interface{} `related_commands:"app, push"`
}

func (StacksCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (StacksCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
