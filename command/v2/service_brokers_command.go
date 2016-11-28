package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type ServiceBrokersCommand struct {
	usage           interface{} `usage:"CF_NAME service-brokers"`
	relatedCommands interface{} `related_commands:"delete-service-broker, disable-service-access, enable-service-access"`
}

func (_ ServiceBrokersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ServiceBrokersCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
