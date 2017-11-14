package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnbindStagingSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-staging-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}        `related_commands:"apps, restart, staging-security-groups"`
}

func (UnbindStagingSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnbindStagingSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
