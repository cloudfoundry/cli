package v6

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ServiceAuthTokensCommand struct {
	usage interface{} `usage:"CF_NAME service-auth-tokens"`
	UI    command.UI
}

func (cmd *ServiceAuthTokensCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	return nil
}

func (cmd *ServiceAuthTokensCommand) Execute(args []string) error {
	cmd.UI.DisplayDeprecationWarning()
	return translatableerror.UnrefactoredCommandError{}
}
