package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetSpaceQuotaCommand struct {
	RequiredArgs    flag.SetSpaceQuotaArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME set-space-quota SPACE_NAME SPACE_QUOTA_NAME"`
	relatedCommands interface{}            `related_commands:"space, space-quotas, spaces"`
}

func (SetSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SetSpaceQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
