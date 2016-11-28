package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type ServiceAccessCommand struct {
	Broker          string      `short:"b" description:"Access for plans of a particular broker"`
	Service         string      `short:"e" description:"Access for service name of a particular service offering"`
	Organization    string      `short:"o" description:"Plans accessible by a particular organization"`
	usage           interface{} `usage:"CF_NAME service-access [-b BROKER] [-e SERVICE] [-o ORG]"`
	relatedCommands interface{} `related_commands:"marketplace, disable-service-access, enable-service-access, service-brokers"`
}

func (_ ServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ServiceAccessCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
