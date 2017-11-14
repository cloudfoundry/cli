package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RouterGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME router-groups"`
	relatedCommands interface{} `related_commands:"create-domain, domains"`
}

func (RouterGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RouterGroupsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
