package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetStagingEnvironmentVariableGroupCommand struct {
	RequiredArgs    flag.ParamsAsJSON `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-staging-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}       `related_commands:"set-env, staging-environment-variable-group"`
}

func (SetStagingEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SetStagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
