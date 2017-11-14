package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteSpaceQuotaCommand struct {
	RequiredArgs    flag.SpaceQuota `positional-args:"yes"`
	Force           bool            `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}     `usage:"CF_NAME delete-space-quota SPACE_QUOTA_NAME [-f]"`
	relatedCommands interface{}     `related_commands:"space-quotas"`
}

func (DeleteSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteSpaceQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
