package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type CreateOrgCommand struct{}

func (_ CreateOrgCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
