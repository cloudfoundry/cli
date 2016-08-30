package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateServiceBrokerCommand struct {
	RequiredArgs    flags.ServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"`
	relatedCommands interface{}             `related_commands:"rename-service-broker, service-brokers"`
}

func (_ UpdateServiceBrokerCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UpdateServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
