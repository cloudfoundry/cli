package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SetEnvCommand struct {
	RequiredArgs    flag.SetEnvironmentArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"`
	relatedCommands interface{}             `related_commands:"apps, env, restart, set-staging-environment-variable-group, set-running-environment-variable-group, unset-env"`
}

func (SetEnvCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SetEnvCommand) Execute(args []string) error {
	//TODO: Be sure to sanitize the WorkAroundPrefix
	return translatableerror.UnrefactoredCommandError{}
}
