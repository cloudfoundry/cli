package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetEnvCommand struct {
	RequiredArgs flags.SetEnvironmentArgs `positional-args:"yes"`
}

func (_ SetEnvCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
