package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DomainsCommand struct {
	usage           interface{} `usage:"CF_NAME domains"`
	relatedCommands interface{} `related_commands:"router-groups, create-route, routes"`
}

func (DomainsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DomainsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
