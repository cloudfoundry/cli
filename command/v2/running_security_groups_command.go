package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RunningSecurityGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME running-security-groups"`
	relatedCommands interface{} `related_commands:"bind-running-security-group, security-group, unbind-running-security-group"`
}

func (RunningSecurityGroupsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RunningSecurityGroupsCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
