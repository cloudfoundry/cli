package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type EnvCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME env APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, set-env, unset-env, running-environment-variable-group, staging-environment-variable-group"`
}

func (EnvCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (EnvCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
