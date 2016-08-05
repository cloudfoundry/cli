package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SetHealthCheckCommand struct {
	RequiredArgs flags.SetHealthCheckArgs `positional-args:"yes"`
}

func (_ SetHealthCheckCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
