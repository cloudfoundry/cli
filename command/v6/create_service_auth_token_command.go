package v6

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"`
	UI           command.UI
}

func (cmd *CreateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	return nil
}

func (cmd *CreateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.UI.DisplayDeprecationWarning()
	return translatableerror.UnrefactoredCommandError{}
}
