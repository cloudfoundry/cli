package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ServiceAuthTokensCommand struct {
	usage interface{} `usage:"CF_NAME service-auth-tokens"`
}

func (ServiceAuthTokensCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ServiceAuthTokensCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
