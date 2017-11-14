package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME security-group SECURITY_GROUP"`
	relatedCommands interface{}        `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group"`
}

func (SecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
