package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UpdateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME update-service-auth-token LABEL PROVIDER TOKEN"`
}

func (UpdateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UpdateServiceAuthTokenCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
