package v6

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteServiceAuthTokenCommand struct {
	RequiredArgs flag.DeleteServiceAuthTokenArgs `positional-args:"yes"`
	Force        bool                            `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}                     `usage:"CF_NAME delete-service-auth-token LABEL PROVIDER [-f]"`
	UI           command.UI
}

func (cmd *DeleteServiceAuthTokenCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	return nil
}

func (cmd *DeleteServiceAuthTokenCommand) Execute(args []string) error {
	cmd.UI.DisplayDeprecationWarning()
	return translatableerror.UnrefactoredCommandError{}
}
