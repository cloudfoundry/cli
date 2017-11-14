package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RenameServiceBrokerCommand struct {
	RequiredArgs    flag.RenameServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"`
	relatedCommands interface{}                  `related_commands:"service-brokers, update-service-broker"`
}

func (RenameServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RenameServiceBrokerCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
