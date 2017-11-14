package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SpaceQuotaCommand struct {
	RequiredArgs flag.SpaceQuota `positional-args:"yes"`
	usage        interface{}     `usage:"CF_NAME space-quota SPACE_QUOTA_NAME"`
}

func (SpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SpaceQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
