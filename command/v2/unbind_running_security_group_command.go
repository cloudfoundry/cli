package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnbindRunningSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-running-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}        `related_commands:"apps, restart, running-security-groups"`
}

func (UnbindRunningSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnbindRunningSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
