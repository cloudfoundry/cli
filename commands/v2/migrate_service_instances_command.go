package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type MigrateServiceInstancesCommand struct {
	RequiredArgs flags.MigrateServiceInstancesArgs `positional-args:"yes"`
	Force        bool                              `short:"f" description:"Force migration without confirmation"`
}

func (_ MigrateServiceInstancesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
