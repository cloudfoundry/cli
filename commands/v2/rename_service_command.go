package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RenameServiceCommand struct {
	RequiredArgs flags.RenameServiceArgs `positional-args:"yes"`
	usage        interface{}             `usage:"CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"`
}

func (_ RenameServiceCommand) Setup(config commands.Config) error {
	return nil
}

func (_ RenameServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
