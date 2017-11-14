package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UpdateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"`
	relatedCommands interface{}            `related_commands:"rename-service-broker, service-brokers"`
}

func (UpdateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UpdateServiceBrokerCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
