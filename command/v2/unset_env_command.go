package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsetEnvCommand struct {
	usage           interface{} `usage:"CF_NAME unset-env APP_NAME ENV_VAR_NAME"`
	relatedCommands interface{} `related_commands:"apps, env, restart, set-staging-environment-variable-group, set-running-environment-variable-group"`
}

func (UnsetEnvCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnsetEnvCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
