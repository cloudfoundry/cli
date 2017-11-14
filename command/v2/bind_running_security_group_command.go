package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type BindRunningSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME bind-running-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}        `related_commands:"apps, bind-security-group, bind-staging-security-group, restart, running-security-groups, security-groups"`
}

func (BindRunningSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (BindRunningSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
