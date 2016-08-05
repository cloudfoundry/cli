package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteServiceBrokerCommand struct {
	RequiredArgs flags.ServiceBroker `positional-args:"yes"`
	Force        bool                `short:"f" description:"Force deletion without confirmation"`
}

func (_ DeleteServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
