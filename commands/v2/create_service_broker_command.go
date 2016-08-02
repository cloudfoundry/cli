package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type CreateServiceBrokerCommand struct{}

func (_ CreateServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
