package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type RunningEnvironmentVariableGroupCommand struct {
	usage           interface{} `usage:"CF_NAME running-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, staging-environment-variable-group"`
}

func (RunningEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (RunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
