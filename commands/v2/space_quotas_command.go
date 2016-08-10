package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type SpaceQuotasCommand struct {
	usage interface{} `usage:"CF_NAME space-quotas"`
}

func (_ SpaceQuotasCommand) Setup() error {
	return nil
}

func (_ SpaceQuotasCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
