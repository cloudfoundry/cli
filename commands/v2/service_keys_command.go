package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ServiceKeysCommand struct {
	RequiredArgs flags.ServiceInstance `positional-args:"yes"`
	usage        interface{}           `usage:"CF_NAME service-keys SERVICE_INSTANCE\n\nEXAMPLES:\n    CF_NAME service-keys mydb"`
}

func (_ ServiceKeysCommand) Setup(config commands.Config) error {
	return nil
}

func (_ ServiceKeysCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
