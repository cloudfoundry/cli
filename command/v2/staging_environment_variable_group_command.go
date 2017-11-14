package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type StagingEnvironmentVariableGroupCommand struct {
	usage           interface{} `usage:"CF_NAME staging-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, running-environment-variable-group"`
}

func (StagingEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (StagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
