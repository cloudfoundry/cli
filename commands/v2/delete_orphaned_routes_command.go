package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type DeleteOrphanedRoutesCommand struct {
	Force bool        `short:"f" description:"Force deletion without confirmation"`
	usage interface{} `usage:"CF_NAME delete-orphaned-routes [-f]"`
}

func (_ DeleteOrphanedRoutesCommand) Setup() error {
	return nil
}

func (_ DeleteOrphanedRoutesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
