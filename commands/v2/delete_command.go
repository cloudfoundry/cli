package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteCommand struct {
	RequiredArgs       flags.AppName `positional-args:"yes"`
	ForceDelete        bool          `short:"f" description:"Force deletion without confirmation"`
	DeleteMappedRoutes bool          `short:"r" description:"Also delete any mapped routes"`
}

func (_ DeleteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
