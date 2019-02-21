package v6

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UpdateServiceAuthTokenCommand struct {
	RequiredArgs flag.ServiceAuthTokenArgs `positional-args:"yes"`
	usage        interface{}               `usage:"CF_NAME update-service-auth-token LABEL PROVIDER TOKEN"`
	UI           command.UI
}

func (cmd *UpdateServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	return nil
}

func (cmd *UpdateServiceAuthTokenCommand) Execute(args []string) error {
	cmd.UI.DisplayDeprecationWarning()
	return translatableerror.UnrefactoredCommandError{}
}
