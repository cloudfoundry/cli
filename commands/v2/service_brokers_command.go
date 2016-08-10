package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type ServiceBrokersCommand struct {
	usage interface{} `usage:"CF_NAME service-brokers"`
}

func (_ ServiceBrokersCommand) Setup() error {
	return nil
}

func (_ ServiceBrokersCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
