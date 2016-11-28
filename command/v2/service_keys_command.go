package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
)

type ServiceKeysCommand struct {
	RequiredArgs    flags.ServiceInstance `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME service-keys SERVICE_INSTANCE\n\nEXAMPLES:\n   CF_NAME service-keys mydb"`
	relatedCommands interface{}           `related_commands:"delete-service-key"`
}

func (_ ServiceKeysCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ServiceKeysCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
