package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RenameSpaceCommand struct {
	RequiredArgs flag.RenameSpaceArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME rename-space SPACE NEW_SPACE"`
}

func (_ RenameSpaceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ RenameSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
