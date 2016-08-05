package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameServiceBrokerCommand struct {
	RequiredArgs flags.RenameServiceBrokerArgs `positional-args:"yes"`
}

func (_ RenameServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
