package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type QuotaCommand struct {
	RequiredArgs    flag.Quota  `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME quota QUOTA"`
	relatedCommands interface{} `related_commands:"org, quotas"`
}

func (QuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (QuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
