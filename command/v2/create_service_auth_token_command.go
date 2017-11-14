package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"`
}

func (CreateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateServiceAuthTokenCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
