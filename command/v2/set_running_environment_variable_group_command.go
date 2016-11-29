package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetRunningEnvironmentVariableGroupCommand struct {
	RequiredArgs    flag.ParamsAsJSON `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-running-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}       `related_commands:"set-env, running-environment-variable-group"`
}

func (_ SetRunningEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SetRunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
