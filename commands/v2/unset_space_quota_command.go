package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type UnsetSpaceQuotaCommand struct{}

func (_ UnsetSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
