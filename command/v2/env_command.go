package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type EnvCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME env APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, set-env, unset-env, running-environment-variable-group, staging-environment-variable-group"`
}

func (_ EnvCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ EnvCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
