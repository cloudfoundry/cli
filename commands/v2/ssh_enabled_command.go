package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type SSHEnabledCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
}

func (_ SSHEnabledCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
