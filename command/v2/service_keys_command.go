package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ServiceKeysCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME service-keys SERVICE_INSTANCE\n\nEXAMPLES:\n   CF_NAME service-keys mydb"`
	relatedCommands interface{}          `related_commands:"delete-service-key"`
}

func (ServiceKeysCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ServiceKeysCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
