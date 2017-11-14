package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type QuotasCommand struct {
	usage interface{} `usage:"CF_NAME quotas"`
}

func (QuotasCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (QuotasCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
