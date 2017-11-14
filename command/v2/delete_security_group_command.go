package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}        `usage:"CF_NAME delete-security-group SECURITY_GROUP [-f]"`
	relatedCommands interface{}        `related_commands:"security-groups"`
}

func (DeleteSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (DeleteSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
