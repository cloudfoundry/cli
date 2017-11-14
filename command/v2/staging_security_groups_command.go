package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type StagingSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME staging-security-groups"`
	relatedCommands interface{} `related_commands:"bind-staging-security-group, security-group, unbind-staging-security-group"`
}

func (StagingSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (StagingSecurityGroupsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
