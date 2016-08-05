package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetRunningEnvironmentVariableGroupCommand struct {
	RequiredArgs flags.ParamsAsJSON `positional-args:"yes"`
}

func (_ SetRunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
