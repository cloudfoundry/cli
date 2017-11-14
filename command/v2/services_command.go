package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ServicesCommand struct {
	usage           interface{} `usage:"CF_NAME services"`
	relatedCommands interface{} `related_commands:"create-service, marketplace"`
}

func (ServicesCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ServicesCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
