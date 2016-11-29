package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateServiceBrokerCommand struct {
	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"`
	relatedCommands interface{}            `related_commands:"rename-service-broker, service-brokers"`
}

func (_ UpdateServiceBrokerCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UpdateServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
