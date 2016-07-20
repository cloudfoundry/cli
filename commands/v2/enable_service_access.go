package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type EnableServiceAccessCommand struct{}

func (_ EnableServiceAccessCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
