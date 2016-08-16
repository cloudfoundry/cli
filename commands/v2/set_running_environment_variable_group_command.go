package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetRunningEnvironmentVariableGroupCommand struct {
	RequiredArgs flags.ParamsAsJSON `positional-args:"yes"`
	usage        interface{}        `usage:"CF_NAME set-running-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
}

func (_ SetRunningEnvironmentVariableGroupCommand) Setup(config commands.Config) error {
	return nil
}

func (_ SetRunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
