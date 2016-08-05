package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceKeysCommand struct {
	RequiredArgs flags.ServiceInstance `positional-args:"yes"`
}

func (_ ServiceKeysCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
