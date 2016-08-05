package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateServiceBrokerCommand struct {
	RequiredArgs flags.ServiceBrokerArgs `positional-args:"yes"`
}

func (_ UpdateServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
