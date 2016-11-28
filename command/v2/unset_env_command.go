package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type UnsetEnvCommand struct {
	usage           interface{} `usage:"CF_NAME unset-env APP_NAME ENV_VAR_NAME"`
	relatedCommands interface{} `related_commands:"apps, env, restart, set-staging-environment-variable-group, set-running-environment-variable-group"`
}

func (_ UnsetEnvCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnsetEnvCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
