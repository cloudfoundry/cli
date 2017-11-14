package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                   `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	usage           interface{}            `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"`
	relatedCommands interface{}            `related_commands:"enable-service-access, service-brokers, target"`
}

func (CreateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateServiceBrokerCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
