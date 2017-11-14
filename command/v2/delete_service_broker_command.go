package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBroker `positional-args:"yes"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}        `usage:"CF_NAME delete-service-broker SERVICE_BROKER [-f]"`
	relatedCommands interface{}        `related_commands:"delete-service, purge-service-offering, service-brokers"`
}

func (DeleteServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteServiceBrokerCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
