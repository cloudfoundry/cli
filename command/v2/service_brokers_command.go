package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ServiceBrokersCommand struct {
	usage           interface{} `usage:"CF_NAME service-brokers"`
	relatedCommands interface{} `related_commands:"delete-service-broker, disable-service-access, enable-service-access"`
}

func (ServiceBrokersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ServiceBrokersCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
