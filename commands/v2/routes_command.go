package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type RoutesCommand struct {
	OrgLevel bool        `long:"orglevel" description:"List all the routes for all spaces of current organization"`
	usage    interface{} `usage:"CF_NAME routes [--orglevel]"`
}

func (_ RoutesCommand) Setup() error {
	return nil
}

func (_ RoutesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
