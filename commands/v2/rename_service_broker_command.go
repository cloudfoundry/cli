package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameServiceBrokerCommand struct {
	RequiredArgs    flags.RenameServiceBrokerArgs `positional-args:"yes"`
	usage           interface{}                   `usage:"CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"`
	relatedCommands interface{}                   `related_commands:"service-brokers, update-service-broker"`
}

func (_ RenameServiceBrokerCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RenameServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
