package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateServiceBrokerCommand struct {
	RequiredArgs    flags.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                    `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	usage           interface{}             `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"`
	relatedCommands interface{}             `related_commands:"enable-service-access, service-brokers, target"`
}

func (_ CreateServiceBrokerCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ CreateServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
