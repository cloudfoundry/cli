package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetQuotaCommand struct {
	RequiredArgs    flag.SetOrgQuotaArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME set-quota ORG QUOTA\n\nTIP:\n   View allowable quotas with 'CF_NAME quotas'"`
	relatedCommands interface{}          `related_commands:"orgs, quotas"`
}

func (SetQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SetQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
