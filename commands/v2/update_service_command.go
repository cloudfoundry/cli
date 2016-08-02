package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type UpdateServiceCommand struct{}

func (_ UpdateServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
