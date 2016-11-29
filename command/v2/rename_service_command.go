package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameServiceCommand struct {
	RequiredArgs    flag.RenameServiceArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"`
	relatedCommands interface{}            `related_commands:"services, update-service"`
}

func (_ RenameServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RenameServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
