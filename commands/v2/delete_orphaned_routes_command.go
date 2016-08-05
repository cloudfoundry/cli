package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type DeleteOrphanedRoutesCommand struct {
	Force bool `short:"f" description:"Force deletion without confirmation"`
}

func (_ DeleteOrphanedRoutesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
