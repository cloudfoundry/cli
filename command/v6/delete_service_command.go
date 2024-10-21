package v6

import (
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/flag"
	"code.cloudfoundry.org/cli/v7/command/translatableerror"
)

type DeleteServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}          `usage:"CF_NAME delete-service SERVICE_INSTANCE [-f]"`
	relatedCommands interface{}          `related_commands:"unbind-service, services"`
}

func (DeleteServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteServiceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
