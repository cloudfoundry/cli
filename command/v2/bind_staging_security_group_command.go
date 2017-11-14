package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type BindStagingSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME bind-staging-security-group SECURITY_GROUP"`
	relatedCommands interface{}        `related_commands:"apps, bind-running-security-group, bind-security-group, restart, security-groups, staging-security-groups"`
}

func (BindStagingSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (BindStagingSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
