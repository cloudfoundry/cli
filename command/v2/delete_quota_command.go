package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteQuotaCommand struct {
	RequiredArgs    flag.Quota  `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"quotas"`
}

func (DeleteQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
