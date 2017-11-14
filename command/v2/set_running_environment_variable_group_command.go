package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetRunningEnvironmentVariableGroupCommand struct {
	RequiredArgs    flag.ParamsAsJSON `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-running-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}       `related_commands:"set-env, running-environment-variable-group"`
}

func (SetRunningEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SetRunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
