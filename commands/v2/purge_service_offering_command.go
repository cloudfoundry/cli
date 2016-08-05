package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type PurgeServiceOfferingCommand struct {
	RequiredArgs flags.Service `positional-args:"yes"`
	Force        bool          `short:"f" description:"Force deletion without confirmation"`
	Provider     string        `short:"p" description:"Provider"`
}

func (_ PurgeServiceOfferingCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
